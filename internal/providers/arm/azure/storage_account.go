package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getARMStorageAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "Microsoft.Storage/storageAccounts",
		RFunc: newARMStorageAccount,
	}
}

func newARMStorageAccount(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	//region := lookupRegion(d, []string{})
	region := "eastus"

	accountKind := "StorageV2"
	if !d.IsEmpty("kind") {
		accountKind = d.Get("kind").String()
	}

	accountReplicationType := d.Get("sku.name").String()
	switch strings.ToLower(accountReplicationType) {
	case "Standard_RAGRS":
		accountReplicationType = "RA-GRS"
	case "Standard_RAGZRS":
		accountReplicationType = "RA-GZRS"
	case "Standard_LRS":
		accountReplicationType = "LRS"
	}

	accountTier := d.Get("sku.tier").String()

	accessTier := "Hot"
	if !d.IsEmpty("properties.accessTier") {
		accessTier = d.Get("properties.accessTier").String()
	}

	nfsv3 := false
	if !d.IsEmpty("properties.isNfsV3Enabled") {
		nfsv3 = d.Get("properties.isNfsV3Enabled").Bool()
	}

	r := &azure.StorageAccount{
		Address:                d.Address,
		Region:                 region,
		AccessTier:             accessTier,
		AccountKind:            accountKind,
		AccountReplicationType: accountReplicationType,
		AccountTier:            accountTier,
		NFSv3:                  nfsv3,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
