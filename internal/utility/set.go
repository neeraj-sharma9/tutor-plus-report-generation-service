package utility

type Set map[int]struct{}

func (s Set) Has(key int) bool {
	_, found := s[key]
	return found
}

func (s Set) AddOrUpdate(v int) {
	s[v] = struct{}{}
}

func (s Set) Remove(v int) {
	delete(s, v)
}

func (s Set) AddMulti(list ...int) {
	for _, v := range list {
		s.AddOrUpdate(v)
	}
}

func (s Set) List() []int {
	var res []int
	for key := range s {
		res = append(res, key)
	}
	return res
}

func (s Set) Join(s2 Set) Set {
	res := Set{}
	for v := range s {
		res.AddOrUpdate(v)
	}

	for v := range s2 {
		res.AddOrUpdate(v)
	}
	return res
}

func (s Set) Difference(s2 Set) Set {
	res := Set{}
	for v := range s {
		if s2.Has(v) {
			continue
		}
		res.AddOrUpdate(v)
	}
	return res
}
