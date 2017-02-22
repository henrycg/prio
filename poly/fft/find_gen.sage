
"""
Sanity checks that g is a generator of order
2^{exp2} of a subgroup of Z*_p.
"""
def is_gen(p, g, exp2):
    if power_mod(g, 2^exp2, p) != 1:
        return False
    for i in range(exp2):
        if power_mod(g, 2^i, p) == 1:
            return False
    return True


"""
Takes as input a prime p such that p-1 has 2^{exp2}
as a factor. Outputs a generator g of order 2^{exp2}.
"""
def find_gen(p, exp2):
    facts = factor(p-1)
    while True:
        r = Zmod(p).random_element()
        for (base, exp) in facts:
            if base == 2: continue
            for e in range(exp):
                r = power_mod(r, base, p)
        if is_gen(p, r, exp2): return r
     
