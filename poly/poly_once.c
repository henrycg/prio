
#include <stdlib.h>
#include <string.h>

#include "poly_batch.h"
#include "poly_once.h"
#include "util.h"


void 
precomp_x_init(precomp_x_t *pre_x,
  const precomp_t *pre, char *xIn)
{
  fmpz_t x;
  fmpz_init_from_gostr(x, xIn);

  fmpz_init(pre_x->modulus);
  fmpz_set(pre_x->modulus, pre->modulus);

  // If the x value at which we want to evaluate is less
  // than n_points, there's no need to do any interpolation,
  // so just short-circuit the computation and return this
  // value.
  //
  // NOTE: This introduces a potential timing attack, since we leak
  // information about x, but in our applications, the probability
  // that x is so short is small.
  const ulong xlong = fmpz_get_ui(x);
  if (xlong < (ulong)pre->n_points) {
    pre_x->short_x = xlong;
  } else {
    pre_x->short_x = -1;
  }

  pre_x->coeffs = safe_malloc(pre->n_points * sizeof(mpz_t));
  pre_x->n_points = pre->n_points;

  for (int i = 0; i < pre_x->n_points; i++)
    fmpz_init(pre_x->coeffs[i]);

  // Given a value x, precompute the coefficients D_i such that:
  //    D_i = C_i * PROD_{i != j} (x - x_j),
  // where the C_i's are stored as s_points in the "pre" struct.
  fmpz_t prod;
  fmpz_t tmp;
  fmpz_init(tmp); 

  fmpz_init(prod); 
  fmpz_set_ui(prod, 1);
  for (int i = 0; i < pre_x->n_points; i++) {
    // Compute tmp = (x - x_i)
    fmpz_sub(tmp, x, pre->x_points[i]);
    fmpz_mul(prod, prod, tmp);
    fmpz_mod(prod, prod, pre->modulus);
  }

  for (int i = 0; i < pre_x->n_points; i++) {
    // Compute PROD_{i != j} (x - x_j)
    fmpz_sub(tmp, x, pre->x_points[i]);
    fmpz_mod(tmp, tmp, pre_x->modulus);
    fmpz_invmod(tmp, tmp, pre->modulus);

    fmpz_mul(tmp, tmp, prod);
    fmpz_mul(tmp, tmp, pre->s_points[i]);
    fmpz_mod(pre_x->coeffs[i], tmp, pre_x->modulus);
  }

  fmpz_clear(tmp); 
  fmpz_clear(prod); 
  fmpz_clear(x);
}

void
precomp_x_clear(precomp_x_t *pre_x)
{
  if (pre_x->short_x >= 0)
    return;

  // Make sure that x-coordinate is not too small
  for (int i = 0; i < pre_x->n_points; i++)
    fmpz_clear(pre_x->coeffs[i]);

  free(pre_x->coeffs);
  fmpz_clear(pre_x->modulus);
}

char *precomp_x_eval(precomp_x_t *pre_x, char **yValues)
{
  if (pre_x->short_x >= 0) {
    char *tmp = yValues[pre_x->short_x];
    char *out = safe_malloc(strlen(tmp) * sizeof(char));
    strcpy(out, tmp);
    return out;
  }

  fmpz_t out;
  fmpz_init(out);
  fmpz_set_ui(out, 0);

  fmpz_t tmp;
  fmpz_init(tmp);
  for (int i = 0; i < pre_x->n_points; i++) {
    fmpz_set_str(tmp, yValues[i], 16);
    fmpz_mul(tmp, tmp, pre_x->coeffs[i]);
    fmpz_mod(tmp, tmp, pre_x->modulus);

    fmpz_add(out, out, tmp);
    fmpz_mod(out, out, pre_x->modulus);
  }
  fmpz_clear(tmp);
  
  char *outstr = fmpz_get_str(NULL, 16, out);
  fmpz_clear(out);
  return outstr;
}

