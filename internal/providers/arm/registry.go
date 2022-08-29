package arm

import (
	"github.com/infracost/infracost/internal/providers/arm/azure"
	"github.com/infracost/infracost/internal/schema"
	"sync"
)

type ResourceRegistryMap map[string]*schema.RegistryItem

var (
	resourceRegistryMap ResourceRegistryMap
	once                sync.Once
)

func GetResourceRegistryMap() *ResourceRegistryMap {
	once.Do(func() {
		resourceRegistryMap = make(ResourceRegistryMap)

		// Merge all resource registries
		for _, registryItem := range azure.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range createFreeResources(azure.FreeResources) {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
	})

	return &resourceRegistryMap
}

func createFreeResources(l []string) []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range l {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:    resourceName,
			NoPrice: true,
			Notes:   []string{"Free resource."},
		})
	}
	return freeResources
}

func GetUsageOnlyResources() []string {
	r := []string{}
	r = append(r, azure.UsageOnlyResources...)
	return r
}
