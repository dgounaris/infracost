package arm

import (
	"fmt"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"strings"
)

type Parser struct {
	ctx *config.ProjectContext
}

func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx}
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				Tags:         d.Tags,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
			}
		}

		res := registryItem.RFunc(d, u)
		if res != nil {
			res.ResourceType = d.Type
			res.Tags = d.Tags
			if u != nil {
				res.EstimationSummary = u.CalcEstimationSummary()
			}
			return res
		}
	}

	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		IsSkipped:    true,
		SkipMessage:  "This resource is not currently supported",
	}
}

func (p *Parser) parseJSON(j []byte, usage map[string]*schema.UsageData) ([]*schema.Resource, []*schema.Resource, error) {
	baseResources := p.loadUsageFileResources(usage)

	if !gjson.ValidBytes(j) {
		return baseResources, baseResources, errors.New("invalid JSON")
	}

	parsed := gjson.ParseBytes(j)

	resources := p.parseJSONResources(baseResources, usage, parsed)

	return nil, resources, nil
}

func (p *Parser) parseJSONResources(baseResources []*schema.Resource, usage map[string]*schema.UsageData, parsed gjson.Result) []*schema.Resource {
	var resources []*schema.Resource
	resources = append(resources, baseResources...)
	resData := p.parseResourceData(parsed)

	for _, d := range resData {
		if r := p.createResource(d, d.UsageData); r != nil {
			resources = append(resources, r)
		}
	}

	return resources
}

func (p *Parser) parseResourceData(planVals gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)

	for _, r := range planVals.Get("resources").Array() {
		t := r.Get("type").String()
		provider := r.Get("provider_name").String()
		addr := r.Get("address").String()

		data := schema.NewResourceData(t, provider, addr, parseTags(r), r)
		data.Metadata = r.Get("infracost_metadata").Map()
		resources[addr] = data

		for _, m := range r.Get("resources").Array() {
			for addr, d := range p.parseResourceData(m) {
				resources[addr] = d
			}
		}
	}

	return resources
}

func parseTags(v gjson.Result) map[string]string {
	tags := make(map[string]string)
	for k, v := range v.Get("tags").Map() {
		tags[k] = v.String()
	}
	return tags
}

func (p *Parser) loadUsageFileResources(u map[string]*schema.UsageData) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for k, v := range u {
		for _, t := range GetUsageOnlyResources() {
			if strings.HasPrefix(k, fmt.Sprintf("%s.", t)) {
				d := schema.NewResourceData(t, "global", k, map[string]string{}, gjson.Result{})
				if r := p.createResource(d, v); r != nil {
					resources = append(resources, r)
				}
			}
		}
	}

	return resources
}
