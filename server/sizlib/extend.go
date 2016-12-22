package sizlib

// (C) Philip Schlump, 2013-2014.

// ExtendStringMap merges map of restuls from multiple JSON reads
func ExtendStringMap(a map[string]string, b map[string]string) map[string]string {
	c := make(map[string]string, len(a)+len(b))
	for i, v := range a {
		c[i] = v
	}
	for i, v := range b {
		c[i] = v
	}
	return c
}

/* vim: set noai ts=4 sw=4: */
