#ifndef _POLY_ONCE__H
#define _POLY_ONCE__H

#include <flint/fmpz.h>

#include "poly_batch.h"

typedef struct {
  int n_points;
  int short_x;
  fmpz_t *coeffs;
  fmpz_t modulus;
} precomp_x_t;

void precomp_x_init(precomp_x_t *pre_x,
  const precomp_t *pre, char *xIn);
void precomp_x_clear(precomp_x_t *pre_x);

char *precomp_x_eval(precomp_x_t *pre_x, char **yValues);

#endif
