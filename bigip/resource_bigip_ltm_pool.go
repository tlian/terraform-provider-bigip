package bigip

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform/helper/schema"
)

var NODE_VALIDATION = regexp.MustCompile(":\\d{2,5}$")

func resourceBigipLtmPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceBigipLtmPoolCreate,
		Read:   resourceBigipLtmPoolRead,
		Update: resourceBigipLtmPoolUpdate,
		Delete: resourceBigipLtmPoolDelete,
		Exists: resourceBigipLtmPoolExists,
		Importer: &schema.ResourceImporter{
			State: resourceBigIpLtmPoolImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the pool",
				ForceNew:     true,
				ValidateFunc: validateF5Name,
			},
			"nodes": &schema.Schema{
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "Nodes to add to the pool. Format node_name:port. e.g. node01:443",
			},

			"monitors": &schema.Schema{
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "Assign monitors to a pool.",
			},

			"allow_nat": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     true,
				Description: "Allow NAT",
			},

			"allow_snat": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     true,
				Description: "Allow SNAT",
			},

			"load_balancing_mode": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "round-robin",
				Description: "Possible values: round-robin, ...",
			},

			"slow_ramp_time": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Slow ramp time for pool members",
			},

			"service_down_action": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "none",
				Description: "Possible values: none, reset, reselect, drop",
			},

			"reselect_tries": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Number of times the system tries to select a new pool member after a failure.",
			},
		},
	}
}

func resourceBigipLtmPoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Get("name").(string)
	d.SetId(name)
	log.Println("[INFO] Creating pool " + name)
	err := client.CreatePool(name)
	if err != nil {
		return err
	}

	err = resourceBigipLtmPoolUpdate(d, meta)
	if err != nil {
		client.DeletePool(name)
		return err
	}

	return resourceBigipLtmPoolRead(d, meta)
}

func resourceBigipLtmPoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	d.Set("name", name)
	log.Println("[INFO] Reading pool " + name)

	pool, err := client.GetPool(name)
	if err != nil {
		return err
	}
	if pool == nil {
		log.Printf("[WARN] Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	nodes, err := client.PoolMembers(name)
	if err != nil {
		return err
	}

	if nodes == nil {
		log.Printf("[WARN] Pool Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	nodeNames := make([]string, 0, len(nodes.PoolMembers))

	for _, node := range nodes.PoolMembers {
		nodeNames = append(nodeNames, node.FullPath)
	}
	if err := d.Set("allow_nat", pool.AllowNAT); err != nil {
		return fmt.Errorf("[DEBUG] Error saving AllowNAT to state for Pool  (%s): %s", d.Id(), err)
	}
	if err := d.Set("allow_snat", pool.AllowSNAT); err != nil {
		return fmt.Errorf("[DEBUG] Error saving AllowSNAT to state for Pool  (%s): %s", d.Id(), err)
	}
	if err := d.Set("load_balancing_mode", pool.LoadBalancingMode); err != nil {
		return fmt.Errorf("[DEBUG] Error saving LoadBalancingMode to state for Pool  (%s): %s", d.Id(), err)
	}
	if err := d.Set("nodes", makeStringSet(&nodeNames)); err != nil {
		return fmt.Errorf("[DEBUG] Error saving Nodes to state for Pool  (%s): %s", d.Id(), err)
	}
	if err := d.Set("slow_ramp_time", pool.SlowRampTime); err != nil {
		return fmt.Errorf("[DEBUG] Error saving SlowRampTime to state for Pool  (%s): %s", d.Id(), err)
	}
	if err := d.Set("service_down_action", pool.ServiceDownAction); err != nil {
		return fmt.Errorf("[DEBUG] Error saving ServiceDownAction to state for Pool  (%s): %s", d.Id(), err)
	}
	if err := d.Set("reselect_tries", pool.ReselectTries); err != nil {
		return fmt.Errorf("[DEBUG] ERror saving ReselectTries to state for Pool  (%s): %s", d.Id(), err)
	}

	monitors := strings.Split(strings.TrimSpace(pool.Monitor), " and ")
	if err := d.Set("monitors", makeStringSet(&monitors)); err != nil {
		return fmt.Errorf("[DEBUG] Error saving Monitors to state for Pool  (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceBigipLtmPoolExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Println("[INFO]   Checking pool " + name + " exists.")

	pool, err := client.GetPool(name)
	if err != nil {
		return false, err
	}

	if pool == nil {
		d.SetId("")
	}

	return pool != nil, nil
}

func resourceBigipLtmPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	//monitors
	var monitors []string
	if m, ok := d.GetOk("monitors"); ok {
		for _, monitor := range m.(*schema.Set).List() {
			monitors = append(monitors, monitor.(string))
		}
	}

	pool := &bigip.Pool{
		AllowNAT:          d.Get("allow_nat").(string),
		AllowSNAT:         d.Get("allow_snat").(string),
		LoadBalancingMode: d.Get("load_balancing_mode").(string),
		SlowRampTime:      d.Get("slow_ramp_time").(int),
		ServiceDownAction: d.Get("service_down_action").(string),
		ReselectTries:     d.Get("reselect_tries").(int),
		Monitor:           strings.Join(monitors, " and "),
	}

	err := client.ModifyPool(name, pool)
	if err != nil {
		return err
	}

	//members
	nodes, err := client.PoolMembers(name)
	if err != nil {
		return err
	}

	nodeNames := make([]string, 0, len(nodes.PoolMembers))

	for _, node := range nodes.PoolMembers {
		nodeNames = append(nodeNames, node.Name)
	}

	existing := makeStringSet(&nodeNames)
	incoming := d.Get("nodes").(*schema.Set)
	delete := existing.Difference(incoming)
	add := incoming.Difference(existing)
	if delete.Len() > 0 {
		for _, d := range delete.List() {
			client.DeletePoolMember(name, d.(string))
		}
	}
	if add.Len() > 0 {
		for _, d := range add.List() {
			client.AddPoolMember(name, d.(string))
		}
	}

	return resourceBigipLtmPoolRead(d, meta)
}

func resourceBigipLtmPoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Deleting pool " + name)

	err := client.DeletePool(name)
	if err != nil {
		return err
	}
	if err == nil {
		log.Printf("[WARN] Pool  (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	return nil
}

func resourceBigIpLtmPoolImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
