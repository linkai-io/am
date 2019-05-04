package differ_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/linkai-io/am/pkg/differ"
)

func TestDiffURL(t *testing.T) {
	expected := "https://google.com/blah/q?asdf="
	u1 := "https://google.com/blah/q?asdf=asdf"
	u2 := "https://google.com/blah/q?asdf=blah2"
	d := differ.New()
	result := d.DiffPatchURL(u1, u2)
	if result != expected {
		t.Fatalf("error expected %s got %s", expected, result)
	}

	expected = "https://google.com//q?asdf="
	u1 = "https://google.com/blah/q?asdf=asdf"
	u2 = "https://google.com/asdf/q?asdf=blah2"
	d = differ.New()
	result = d.DiffPatchURL(u1, u2)
	if result != expected {
		t.Fatalf("error expected %s got %s", expected, result)
	}
}

func TestDiffMicrosoft1(t *testing.T) {
	f1, err := ioutil.ReadFile("testdata/dom1.html")
	if err != nil {
		t.Fatalf("error reading file 1 %v\n", err)
	}

	f2, err := ioutil.ReadFile("testdata/dom2.html")
	if err != nil {
		t.Fatalf("error reading file 2 %v\n", err)
	}
	d := differ.New()
	ctx := context.Background()
	hash, same := d.DiffHash(ctx, string(f1), string(f2))
	if !same {
		t.Fatalf("error not same")
	}
	t.Logf("%s\n", hash)
}

func TestDiffMicrosoft2(t *testing.T) {
	f1, err := ioutil.ReadFile("testdata/dom3.html")
	if err != nil {
		t.Fatalf("error reading file 1 %v\n", err)
	}

	f2, err := ioutil.ReadFile("testdata/dom4.html")
	if err != nil {
		t.Fatalf("error reading file 2 %v\n", err)
	}
	d := differ.New()
	ctx := context.Background()
	hash, same := d.DiffHash(ctx, string(f1), string(f2))
	if !same {
		t.Fatalf("error not same")
	}
	t.Logf("%s\n", hash)
}

func TestDiffNCC(t *testing.T) {
	f1, err := ioutil.ReadFile("testdata/dom5.html")
	if err != nil {
		t.Fatalf("error reading file 1 %v\n", err)
	}

	f2, err := ioutil.ReadFile("testdata/dom6.html")
	if err != nil {
		t.Fatalf("error reading file 2 %v\n", err)
	}
	d := differ.New()
	ctx := context.Background()
	hash, same := d.DiffHash(ctx, string(f1), string(f2))
	if !same {
		t.Fatalf("error not same")
	}
	t.Logf("%s\n", hash)
}
func TestDiffGoogle(t *testing.T) {
	f1, err := ioutil.ReadFile("testdata/dom7.html")
	if err != nil {
		t.Fatalf("error reading file 1 %v\n", err)
	}

	f2, err := ioutil.ReadFile("testdata/dom8.html")
	if err != nil {
		t.Fatalf("error reading file 2 %v\n", err)
	}
	d := differ.New()
	ctx := context.Background()
	hash, same := d.DiffHash(ctx, string(f1), string(f2))
	if !same {
		t.Fatalf("error not same")
	}
	t.Logf("%s\n", hash)
}

func TestDiffSimple(t *testing.T) {
	f1 := `<script>
	_pageBITags={"pageTags":{"uri":"https://www.microsoft.com/ja-jp/","mkt":"ja-jp","referrerUri":"","browserGroup":"uplevel.web.pc.webkit.chrome","expId":"EX:sfwaab","enabledFeatures":"optimizely_disabled:1,uhf_retailstore2:1,UhfPb:1,UhfUsePh:1,EnableLocaleDetection:1,UhfSwp:1,uhfgreenid:1,enable_sasslib_minification_runtime:1,core_cookiecompliance_enabled:1,core_akamai_im_enabled:1,coreui_hero_image_resize_90:1,uhf_as_iris:1,core_use_coreui_mwf:1,coreui_makeimagebackgroundtransparent:1,f_audiencemanager_disabled:1,core_BypassJWTValidation:1,MSADisableForceSignin:1,IsRtoRuleDisabled:1,DisableToSkipMarketdetectionforUknownRoutes:1,f_video_uselegacyservice:1,uhf_magic_triangle:1,RelevanceOverride:1,pipeline:1,coreui_videomodule_useflexsize:1,EnableAzureActiveDirectory20:1,AutoCORS_disabled:1,IsIrisV4Enabled:1,f_video_useadaptive:1,core_uhf_access_policy:1,uhf_st_enabled:1,jquery_latest:1,core_trustedCors:1,DisableOneRFSearchRoute:1,clientTypeSfw:1,ResolveDataProviderByPartnerNameSpace:1,core_disable_extensibility:1,InvokeLoginAuthorizeAndRedirect:1,AllowIncludeExclusivityArguments:1,uhf_stick_footer_to_bottom:1,EnableFetchOfKnownDocument:1,boomerang_disabled:1,DisableTATToken:1,retailServerFromTenantConfig:1,node_scnr_blob:1,node_disable_app_cache:1,core_pageTypeToken:1","signedInStatus":false,"pv":"0.1","dv":"2019/04/18 22:41:39 +00:00","jsEnabled":true,"isTented":false,"isCached":false,"isOneRf":true,"isCorpNet":false,"isStatic":false,"tags":{"serviceName":"marketingsites-prod-odwestcentralus"},"shareAuthStatus":false,"userConsent":true,"muidDomain":"microsoft.com","autoCapture":{"pageView":true,"onLoad":true,"click":true,"scroll":true,"resize":true,"context":true,"jsError":true,"perf":true},"scripts":"JQuery,Comscore,AudienceManager","disableJsll":false,"pageTemplateId":"RE2MDAF","pageSubType":"RE2MDAF","canvasType":"Web","deviceFamily":null,"authType":null,"appId":"MicrosoftHP","tasId":"791b7d85-79bb-49c6-8c35-947da117e29d","tasMuid":"3135F845EDBC64AB0C8BF501EC94657A","pageName":"Homepage","pageType":"HP.AllModules","env":"onerf_prod","cV":"XCxD6yy/GECOczUl.0","imprGuid":"791b7d85-79bb-49c6-8c35-947da117e29d"},"elementTag":"data-m","defaultParent":"Body","defaultValue":"Unspecified"};
</script>`
	f2 := `<script>
	_pageBITags={"pageTags":{"uri":"https://www.microsoft.com/ja-jp/","mkt":"ja-jp","referrerUri":"","browserGroup":"uplevel.web.pc.webkit.chrome","expId":"EX:sfwaaa","enabledFeatures":"optimizely_disabled:1,uhf_retailstore2:1,UhfPb:1,UhfUsePh:1,EnableLocaleDetection:1,UhfSwp:1,uhfgreenid:1,enable_sasslib_minification_runtime:1,core_cookiecompliance_enabled:1,core_akamai_im_enabled:1,coreui_hero_image_resize_90:1,uhf_as_iris:1,core_use_coreui_mwf:1,coreui_makeimagebackgroundtransparent:1,f_audiencemanager_disabled:1,core_BypassJWTValidation:1,MSADisableForceSignin:1,IsRtoRuleDisabled:1,DisableToSkipMarketdetectionforUknownRoutes:1,f_video_uselegacyservice:1,uhf_magic_triangle:1,RelevanceOverride:1,pipeline:1,coreui_videomodule_useflexsize:1,EnableAzureActiveDirectory20:1,AutoCORS_disabled:1,IsIrisV4Enabled:1,f_video_useadaptive:1,core_uhf_access_policy:1,uhf_st_enabled:1,jquery_latest:1,core_trustedCors:1,DisableOneRFSearchRoute:1,clientTypeSfw:1,ResolveDataProviderByPartnerNameSpace:1,core_disable_extensibility:1,InvokeLoginAuthorizeAndRedirect:1,AllowIncludeExclusivityArguments:1,uhf_stick_footer_to_bottom:1,EnableFetchOfKnownDocument:1,boomerang_disabled:1,DisableTATToken:1,retailServerFromTenantConfig:1,node_scnr_blob:1,node_disable_app_cache:1,core_pageTypeToken:1","signedInStatus":false,"pv":"0.1","dv":"2019/04/18 22:41:39 +00:00","jsEnabled":true,"isTented":false,"isCached":false,"isOneRf":true,"isCorpNet":false,"isStatic":false,"tags":{"serviceName":"marketingsites-prod-odeastus"},"shareAuthStatus":false,"userConsent":true,"muidDomain":"microsoft.com","autoCapture":{"pageView":true,"onLoad":true,"click":true,"scroll":true,"resize":true,"context":true,"jsError":true,"perf":true},"scripts":"JQuery,Comscore,AudienceManager","disableJsll":false,"pageTemplateId":"RE2MDAF","pageSubType":"RE2MDAF","canvasType":"Web","deviceFamily":null,"authType":null,"appId":"MicrosoftHP","tasId":"8804e6d3-ca9a-46ae-ad19-2c21120ed81a","tasMuid":"18690046B6976CD5217D0D02B76A6D27","pageName":"Homepage","pageType":"HP.AllModules","env":"onerf_prod","cV":"mErK7Z6lGUCpX4w5.0","imprGuid":"8804e6d3-ca9a-46ae-ad19-2c21120ed81a"},"elementTag":"data-m","defaultParent":"Body","defaultValue":"Unspecified"};
</script>`
	d := differ.New()
	ctx := context.Background()
	hash, same := d.DiffHash(ctx, f1, f2)
	if !same {
		t.Fatalf("error not same")
	}
	t.Logf("%s\n", hash)
}
