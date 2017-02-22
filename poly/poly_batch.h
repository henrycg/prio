#ifndef _POLY_BATCH__H
#define _POLY_BATCH__H

#include <flint/fmpz_mod_poly.h>

struct tree_s {
  fmpz_mod_poly_t poly; 
  struct tree_s *left; 
  struct tree_s *right;
};

struct precomp_s {
  int n_points;
  fmpz_t modulus;
  fmpz_t *x_points;
  struct tree_s tree;
  fmpz_mod_poly_t deriv;
  fmpz_t *s_points;
};

typedef struct tree_s tree_t;
typedef struct precomp_s precomp_t;

void poly_batch_precomp_init(struct precomp_s *pre, char *modIn, 
    int n_points, char **pointsXin);
void poly_batch_precomp_clear(struct precomp_s *pre);

void poly_batch_init(fmpz_mod_poly_t poly, struct precomp_s *pre);
void poly_batch_clear(fmpz_mod_poly_t poly);

void poly_batch_interpolate(fmpz_mod_poly_t poly, struct precomp_s *pre, char **pointsYin);

char *poly_batch_evaluate_once(fmpz_mod_poly_t poly, char *xIn);
char *poly_batch_evaluate(fmpz_mod_poly_t poly, int n_points, char **pointsXin);

#endif
