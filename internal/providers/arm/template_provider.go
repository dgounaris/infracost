package arm

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"os"
)

type TemplateProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

func NewTemplateProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &TemplateProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *TemplateProvider) Type() string {
	return "arm"
}

func (p *TemplateProvider) DisplayType() string {
	return "ARM"
}

func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TemplateProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	jsons, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading ARM template JSON file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.ProjectConfig.Name, p.ctx.RunContext.IsCloudEnabled())

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)
	if jsons[0] == 239 && jsons[1] == 187 && jsons[2] == 191 {
		jsons = jsons[3:] // remove BOB if exists
	}
	pastResources, resources, err := parser.parseJSON(jsons, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing ARM template file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	if !p.includePastResources {
		project.PastResources = nil
	}

	return []*schema.Project{project}, nil
}
