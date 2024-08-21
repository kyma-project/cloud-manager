package meta

import "context"

type cceeDomainKeyType struct{}

var cceeDomainKey = cceeDomainKeyType{}

type cceeProjectKeyType struct{}

var cceeProjectKey = cceeProjectKeyType{}

type cceeRegionKeyType struct{}

var cceeRegionKey = cceeProjectKeyType{}

func GetCceeDomain(ctx context.Context) string {
	x := ctx.Value(cceeDomainKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetCceeDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, cceeDomainKey, domain)
}

func GetCceeProject(ctx context.Context) string {
	x := ctx.Value(cceeProjectKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetCceeProject(ctx context.Context, project string) context.Context {
	return context.WithValue(ctx, cceeProjectKey, project)
}

func GetCceeRegion(ctx context.Context) string {
	x := ctx.Value(cceeRegionKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetCceeRegion(ctx context.Context, region string) context.Context {
	return context.WithValue(ctx, cceeRegionKey, region)
}

func SetCeeDomainProjectRegion(ctx context.Context, domain, project, region string) context.Context {
	ctx = SetCceeDomain(ctx, domain)
	ctx = SetCceeProject(ctx, project)
	ctx = SetCceeRegion(ctx, region)
	return ctx
}
