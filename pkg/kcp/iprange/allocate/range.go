package allocate

import (
	"fmt"
	"math/big"
	"net"
)

type rng struct {
	s     string
	n     *net.IPNet
	first *big.Int
	last  *big.Int
}

func parseRange(s string) (*rng, error) {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}

	ipAllFfs := make(net.IP, len(n.IP))
	for i := 0; i < len(ipAllFfs); i++ {
		ipAllFfs[i] = 255
	}

	mask := big.NewInt(0).SetBytes(n.Mask)
	first := big.NewInt(0).SetBytes(n.IP)
	last := big.NewInt(0).Or(
		big.NewInt(0).And(first, mask),
		big.NewInt(0).Xor(mask, big.NewInt(0).SetBytes(ipAllFfs)),
	)

	return &rng{
		s:     s,
		n:     n,
		first: first,
		last:  last,
	}, nil
}

func (r *rng) overlaps(o *rng) bool {
	// r: f-l       f---l     f------l     f-l        f---l         f-l
	// o:     f-l     f---l     f--l     f-----l   f----l      f-l
	//      no        yes       yes        yes        yes         no
	// r_first <= o_last && o_first <= r_last
	return r.first.Cmp(o.last) <= 0 && o.first.Cmp(r.last) <= 0
}

func (r *rng) len() int {
	return int(big.NewInt(0).Sub(r.last, r.first).Int64()) + 1
}

func (r *rng) next() *rng {
	ones, _ := r.n.Mask.Size()
	return r.nextWithOnes(ones)
}

func (r *rng) nextWithOnes(ones int) *rng {
	ip := net.IP(big.NewInt(0).Add(r.last, big.NewInt(1)).Bytes())
	s := fmt.Sprintf("%s/%d", ip.String(), ones)
	res, err := parseRange(s)
	if err != nil {
		return nil
	}
	return res
}
