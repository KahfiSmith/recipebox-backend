# RecipeBox Database (Draft)

Dokumen ini berisi rancangan awal skema database untuk RecipeBox. Ini
masih draft dan dapat berubah sesuai kebutuhan implementasi API.

## Prinsip Umum

- Database: PostgreSQL
- Primary key menggunakan `bigserial`.
- Timestamp menggunakan `timestamptz` (`created_at`, `updated_at`).
- Relasi many-to-many menggunakan tabel junction.
- Pencarian resep berdasarkan bahan menggunakan tabel `recipe_ingredients`.

## Tabel Inti

### recipes

Menyimpan data resep.

- `id` bigserial PK
- `title` text not null
- `description` text null
- `instructions` text null
- `prep_minutes` int null
- `cook_minutes` int null
- `servings` int null
- `created_at` timestamptz not null default now()
- `updated_at` timestamptz not null default now()

Index:
- `recipes_title_idx` on (`title`)

### ingredients

Master bahan.

- `id` bigserial PK
- `name` text not null
- `created_at` timestamptz not null default now()
- `updated_at` timestamptz not null default now()

Index:
- unique `ingredients_name_uq` on (`name`)

### recipe_ingredients

Relasi resep dan bahan (many-to-many) + takaran.

- `id` bigserial PK
- `recipe_id` bigint not null references recipes(id) on delete cascade
- `ingredient_id` bigint not null references ingredients(id) on delete restrict
- `quantity` numeric(10,2) null
- `unit` text null
- `note` text null

Index:
- unique `recipe_ingredients_uq` on (`recipe_id`, `ingredient_id`)
- `recipe_ingredients_recipe_id_idx` on (`recipe_id`)
- `recipe_ingredients_ingredient_id_idx` on (`ingredient_id`)

## Meal Plan

### meal_plans

Header meal plan (harian/mingguan).

- `id` bigserial PK
- `name` text not null
- `start_date` date not null
- `end_date` date not null
- `created_at` timestamptz not null default now()
- `updated_at` timestamptz not null default now()

### meal_plan_items

Item meal plan per tanggal.

- `id` bigserial PK
- `meal_plan_id` bigint not null references meal_plans(id) on delete cascade
- `recipe_id` bigint not null references recipes(id) on delete restrict
- `plan_date` date not null
- `meal_type` text null

Index:
- `meal_plan_items_plan_date_idx` on (`plan_date`)
- `meal_plan_items_meal_plan_id_idx` on (`meal_plan_id`)

## Shopping List

### shopping_lists

Header shopping list yang di-generate.

- `id` bigserial PK
- `name` text null
- `source_meal_plan_id` bigint null references meal_plans(id) on delete set null
- `created_at` timestamptz not null default now()
- `updated_at` timestamptz not null default now()

### shopping_list_items

Item belanja, hasil agregasi bahan dari resep.

- `id` bigserial PK
- `shopping_list_id` bigint not null references shopping_lists(id) on delete cascade
- `ingredient_id` bigint not null references ingredients(id) on delete restrict
- `quantity` numeric(10,2) null
- `unit` text null
- `note` text null
- `is_checked` boolean not null default false

Index:
- `shopping_list_items_shopping_list_id_idx` on (`shopping_list_id`)
- unique `shopping_list_items_uq` on (`shopping_list_id`, `ingredient_id`, `unit`)

## Catatan Implementasi

- Jika butuh pencarian resep berbasis bahan, gunakan join ke
  `recipe_ingredients` dan `ingredients`.
- Untuk multi-user di fase berikutnya, tambahkan tabel `users` dan field
  `user_id` pada tabel yang relevan (recipes, meal_plans, shopping_lists).
