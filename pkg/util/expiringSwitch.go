package util

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/patrickmn/go-cache"
)

var _expiringSwitch ExpiringSwitchType

func init() {
	expiration := 10 * time.Minute
	cleanup := 15 * time.Minute
	_expiringSwitch = &expiringFlagImpl{
		cache:      cache.New(expiration, cleanup),
		expiration: expiration,
	}
}

func ExpiringSwitch() ExpiringSwitchType {
	return _expiringSwitch
}

type ExpiringSwitchType interface {
	Key(data ...string) *ExpiringFlagKey
}

type ExpiringFlagKey struct {
	key        string
	expiration time.Duration
	ec         *cache.Cache
}

func (k *ExpiringFlagKey) IfNotRecently(cb func()) {
	if k.key != "" && k.ec != nil && !k.present() {
		cb()
		k.set()
	}
}

func (k *ExpiringFlagKey) present() bool {
	_, present := k.ec.Get(k.key)
	return present
}

func (k *ExpiringFlagKey) set() {
	k.ec.Set(k.key, nil, k.expiration)
}

type expiringFlagImpl struct {
	cache      *cache.Cache
	expiration time.Duration
}

func (in *expiringFlagImpl) Key(data ...string) *ExpiringFlagKey {
	hasher := sha256.New()
	for _, d := range data {
		hasher.Write([]byte(d))
	}
	hashSum := hasher.Sum(nil)
	hexHash := hex.EncodeToString(hashSum)
	return &ExpiringFlagKey{
		key:        hexHash,
		expiration: in.expiration,
		ec:         in.cache,
	}
}
