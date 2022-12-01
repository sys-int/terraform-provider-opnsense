package unbound

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	opn_core "github.com/sys-int/opnsense-api/api"
	opn_unbound "github.com/sys-int/opnsense-api/api/unbound"
	"sync"
	common "terraform-sysint-os-dns/opnsense/common"
)

var mtx sync.Mutex

func resourceHostOverrideCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "creating unbound override host")
	var diags diag.Diagnostics
	client := meta.(common.IProviderClient)
	tflog.Debug(ctx, fmt.Sprintf("conn url=\"%s\" key=\"%s\" secret=\"%s\"", client.GetConn().BaseUrl.String(), client.GetConn().ApiKey, client.GetConn().ApiSecret))

	api := opn_unbound.UnboundApi{client.GetConn()}

	mtx.Lock()
	host := unmarshalHost(ctx, d, &opn_unbound.HostOverride{})
	uuid, err := api.HostOverrideCreate(*host)
	api.ServiceRestart()
	mtx.Unlock()
	if err != nil {
		tflog.Error(ctx, "error creating override host "+err.Error())
		return diag.FromErr(err)
	}

	d.SetId(uuid)
	resourceHostOverrideRead(ctx, d, meta)

	return diags
}

func resourceHostOverrideRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "read unbound override host")
	var diags diag.Diagnostics
	client := meta.(common.IProviderClient)
	api := opn_unbound.UnboundApi{client.GetConn()}
	host, err := api.HostEntryGetByUuid(d.Id())

	if err != nil {
		switch err.(type) {
		case *opn_core.NotFoundError:
			d.SetId("")
			return nil
		default:
			return diag.FromErr(err)
		}
	} else {
		host.Uuid = d.Id()
		marshalHost(ctx, d, host)
	}
	return diags
}

func resourceHostOverrideDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "delete unbound override host")
	var diags diag.Diagnostics
	client := meta.(common.IProviderClient)
	api := opn_unbound.UnboundApi{client.GetConn()}
	mtx.Lock()
	err := api.HostEntryRemove(d.Id())
	api.ServiceRestart()
	mtx.Unlock()
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceHostOverrideUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "update unbound override host")
	var diags diag.Diagnostics
	client := meta.(common.IProviderClient)
	api := opn_unbound.UnboundApi{client.GetConn()}
	host := unmarshalHost(ctx, d, &opn_unbound.HostOverride{})
	mtx.Lock()
	uuid, err := api.HostOverrideUpdate(*host)
	api.ServiceRestart()
	mtx.Unlock()
	host.Uuid = uuid
	marshalHost(ctx, d, *host)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
