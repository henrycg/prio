
#include <stdio.h>
#include <stdlib.h>

#include "util.h"
#include "poly_batch.h"


static void fast_linear_comp(fmpz_mod_poly_t poly, const struct tree_s *pre,
  int n_points, fmpz_t *pointsX, fmpz_t *pointsY);
static void fast_interpolate(fmpz_mod_poly_t poly, const struct precomp_s *pre, fmpz_t *pointsY);
static void fast_evaluate(fmpz_t *pointsY, const fmpz_mod_poly_t poly, 
    const struct tree_s *pre, int n_points, fmpz_t *pointsX);

static void tree_init(struct tree_s *pre, const fmpz_t modulus, 
    int n_points, fmpz_t *pointsX);
void tree_clear(struct tree_s *pre);

static void c_precomp_init(struct precomp_s *pre, const fmpz_t modulus, 
  int n_points, fmpz_t *pointsX);

void
poly_batch_init(fmpz_mod_poly_t poly, struct precomp_s *pre)
{
  fmpz_mod_poly_init(poly, pre->modulus);
  fmpz_mod_poly_set_coeff_ui(poly, 0, 0);
}

void 
poly_batch_clear(fmpz_mod_poly_t poly)
{
  fmpz_mod_poly_clear(poly);
}

char *
poly_batch_evaluate_once(fmpz_mod_poly_t poly, char *xIn)
{
  fmpz_t x;
  fmpz_init_from_gostr(x, xIn);

  fmpz_t y;
  fmpz_init(y);

  fmpz_mod_poly_evaluate_fmpz(y, poly, x);
  char *out = fmpz_get_str(NULL, 16, y);

  fmpz_clear(x);
  fmpz_clear(y);

  return out;
}

char *
poly_batch_evaluate(fmpz_mod_poly_t poly, int n_points, char **pointsXin)
{
  fmpz_t xs[n_points];
  fmpz_t ys[n_points];
  for (int i = 0; i < n_points; i++) {
    fmpz_init_from_gostr(xs[i], pointsXin[i]);
    fmpz_init(ys[i]);
  }

  fmpz_mod_poly_evaluate_fmpz_vec_fast(ys[0], poly, xs[0], n_points);

  for (int i = 0; i < n_points; i++) {
    fmpz_clear(xs[i]);
  }

  return fmpz_array_to_str(n_points, ys);
}


void
poly_batch_interpolate(fmpz_mod_poly_t poly, struct precomp_s *pre, char **pointsYin)
{
  // We use the algorithm from:
  //    "Modern Computer Algebra"
  //    Joachim von zur Gathen and Jurgen Gerhard
  //    Chapter 10, Algorithm 10.11 
  fmpz_t pointsY[pre->n_points];

  for (int i = 0; i < pre->n_points; i++) {
    fmpz_init_from_gostr(pointsY[i], pointsYin[i]);
  }

  fast_interpolate(poly, pre, pointsY);

  for (int i = 0; i < pre->n_points; i++) {
    fmpz_clear(pointsY[i]);
  }
}

static void
tree_init(struct tree_s *pre, const fmpz_t modulus, 
    int n_points, fmpz_t *pointsX)
{
  // On input (x_1, ..., x_n), compute the polynomial
  //    f(x) = (x - x_1)(x - x_2) ... (x - x_n)
  // using a divide-and-conquer approach, saving the
  // intermediate results in a tree.

  fmpz_mod_poly_init(pre->poly, modulus);
  fmpz_mod_poly_zero(pre->poly);

  if (n_points == 1) {
    pre->left = NULL; 
    pre->right = NULL; 

    fmpz_t tmp;
    fmpz_init(tmp);
    fmpz_set(tmp, pointsX[0]);
    fmpz_neg(tmp, tmp);
  
    // In base case, polynomial is (x - x_i)
    fmpz_mod_poly_set_coeff_ui(pre->poly, 1, 1);
    fmpz_mod_poly_set_coeff_fmpz(pre->poly, 0, tmp);
    fmpz_clear(tmp);
    return;
  }

  pre->left = safe_malloc(sizeof(struct tree_s));
  pre->right = safe_malloc(sizeof(struct tree_s));

  if (!pre->left || !pre->right) {
    fprintf(stderr, "Ran out of memory!\n");
    exit(1);
  }

  const int k = n_points / 2;
  // Compute the left polynomial recursively on the first k points
  tree_init(pre->left, modulus, k, pointsX);

  // Compute the right polynomial on the rest of the points
  tree_init(pre->right, modulus, n_points - k, &pointsX[k]);

  // Store the product
  fmpz_mod_poly_mul(pre->poly, pre->left->poly, pre->right->poly);
}

static void c_precomp_init(struct precomp_s *pre, const fmpz_t modulus, 
  int n_points, fmpz_t *pointsX)
{
  pre->x_points = safe_malloc(n_points * sizeof(fmpz_t));
  pre->s_points = safe_malloc(n_points * sizeof(fmpz_t));

  for (int i = 0; i < n_points; i++) {
    fmpz_init(pre->x_points[i]);
    fmpz_init(pre->s_points[i]);

    fmpz_set(pre->x_points[i], pointsX[i]);
  }

  pre->n_points = n_points;
  tree_init(&pre->tree, modulus, n_points, pointsX);

  // Compute derivative of the roots
  fmpz_mod_poly_init(pre->deriv, modulus);
  fmpz_mod_poly_derivative(pre->deriv, pre->tree.poly);

  // Compute s_i's
  fast_evaluate(pre->s_points, pre->deriv, &pre->tree, n_points, pointsX);

  for(int i = 0; i < n_points; i++) {
    fmpz_invmod(pre->s_points[i], pre->s_points[i], modulus);
  }
}

void 
poly_batch_precomp_init(struct precomp_s *pre, char *modIn, 
    int n_points, char **pointsXin)
{
  fmpz_init_from_gostr(pre->modulus, modIn);

  fmpz_t pointsX[n_points];
  for (int i = 0; i < n_points; i++) {
    fmpz_init_from_gostr(pointsX[i], pointsXin[i]);
  }

  c_precomp_init(pre, pre->modulus, n_points, pointsX);

  for (int i = 0; i < pre->n_points; i++) {
    fmpz_clear(pointsX[i]);
  }
}

void poly_batch_precomp_clear(struct precomp_s *pre)
{
  for (int i = 0; i < pre->n_points; i++) {
    fmpz_clear(pre->s_points[i]);
    fmpz_clear(pre->x_points[i]);
  }

  tree_clear(&pre->tree);
  fmpz_mod_poly_clear(pre->deriv);
}

void tree_clear(struct tree_s *pre) 
{
  if (pre->left) {
    tree_clear(pre->left);
    free(pre->left);
  }
  if (pre->right) {
    tree_clear(pre->right);
    free(pre->right);
  }

  fmpz_mod_poly_clear(pre->poly);
}

static void
fast_interpolate(fmpz_mod_poly_t poly, const struct precomp_s *pre, fmpz_t *pointsY)
{
  const fmpz *mod = fmpz_mod_poly_modulus(poly);

  fmpz_t ys[pre->n_points];

  for(int i = 0; i < pre->n_points; i++) {
    fmpz_init(ys[i]);
    fmpz_mul(ys[i], pre->s_points[i], pointsY[i]);
    fmpz_mod(ys[i], ys[i], mod);
  }

  fast_linear_comp(poly, &pre->tree, pre->n_points, pre->x_points, ys); 

  for(int i = 0; i < pre->n_points; i++)
    fmpz_clear(ys[i]);
}

static void
fast_linear_comp(fmpz_mod_poly_t poly, const struct tree_s *pre,
  int n_points, fmpz_t *pointsX, fmpz_t *pointsY)
{
  // This is Algorithm 10.9 from the book cited above.
  //
  // Input (x_1, ..., x_N)   and   (y_1, ..., y_N)
  // Output:
  //      SUM_i (y_i * m(x))/(x - x_i)
  // where m(x) = (x - x_1)(x - x_2)...(x - x_n) 
  const fmpz *mod = fmpz_mod_poly_modulus(poly);

  if (n_points == 1) {
    fmpz_mod_poly_zero(poly);
    fmpz_mod_poly_set_coeff_fmpz(poly, 0, pointsY[0]);  
    return;
  } 

  const int k = n_points/2;

  fmpz_mod_poly_t r0, r1;
  fmpz_mod_poly_init(r0, mod);
  fmpz_mod_poly_init(r1, mod);

  fast_linear_comp(r0, pre->left, k, pointsX, pointsY);
  fast_linear_comp(r1, pre->right, n_points - k, pointsX + k, pointsY + k);

  fmpz_mod_poly_mul(r0, r0, pre->right->poly); 
  fmpz_mod_poly_mul(r1, r1, pre->left->poly); 

  fmpz_mod_poly_add(poly, r0, r1);

  fmpz_mod_poly_clear(r0);
  fmpz_mod_poly_clear(r1);
}

static void
fast_evaluate(fmpz_t *pointsY, const fmpz_mod_poly_t poly, const struct tree_s *pre, int n_points, fmpz_t *pointsX)
{
  if (n_points == 1) {
    fmpz_mod_poly_get_coeff_fmpz(pointsY[0], poly, 0);
    return;
  }

  const fmpz *mod = fmpz_mod_poly_modulus(poly);
  fmpz_mod_poly_t r0, r1;
  fmpz_mod_poly_init(r0, mod);
  fmpz_mod_poly_init(r1, mod);

  fmpz_mod_poly_rem(r0, poly, pre->left->poly);
  fmpz_mod_poly_rem(r1, poly, pre->right->poly);

  const int k = n_points/2;
  fast_evaluate(pointsY, r0, pre->left, k, pointsX);
  fast_evaluate(pointsY + k, r1, pre->right, n_points - k, pointsX + k);

  fmpz_mod_poly_clear(r0);
  fmpz_mod_poly_clear(r1);
}

