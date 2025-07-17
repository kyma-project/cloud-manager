package meta

import "context"

type sapDomainKeyType struct{}

var sapDomainKey = sapDomainKeyType{}

type sapProjectKeyType struct{}

var sapProjectKey = sapProjectKeyType{}

type sapRegionKeyType struct{}

var sapRegionKey = sapRegionKeyType{}

func GetSapDomain(ctx context.Context) string {
	x := ctx.Value(sapDomainKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetSapDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, sapDomainKey, domain)
}

func GetSapProject(ctx context.Context) string {
	x := ctx.Value(sapProjectKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetSapProject(ctx context.Context, project string) context.Context {
	return context.WithValue(ctx, sapProjectKey, project)
}

func GetSapRegion(ctx context.Context) string {
	x := ctx.Value(sapRegionKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetSapRegion(ctx context.Context, region string) context.Context {
	return context.WithValue(ctx, sapRegionKey, region)
}

func SetSapDomainProjectRegion(ctx context.Context, domain, project, region string) context.Context {
	ctx = SetSapDomain(ctx, domain)
	ctx = SetSapProject(ctx, project)
	ctx = SetSapRegion(ctx, region)
	return ctx
}
