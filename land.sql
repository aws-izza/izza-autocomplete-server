-- auto-generated definition
create table land
(
    id                   bigint generated always as identity
        primary key,
    shape_id             bigint,
    unique_no            varchar(20)             not null
        unique,
    full_code            varchar(10),
    address              varchar(128),
    ledger_division_code smallint,
    ledger_division_name varchar(10),
    base_year            smallint,
    base_month           smallint,
    land_category_code   smallint,
    land_category_name   varchar(20),
    land_area            numeric(12, 2),
    use_district_code1   smallint,
    use_district_name1   varchar(50),
    use_district_code2   smallint,
    use_district_name2   varchar(50),
    land_use_code        smallint,
    land_use_name        varchar(20),
    terrain_height_code  smallint,
    terrain_height_name  varchar(20),
    terrain_shape_code   smallint,
    terrain_shape_name   varchar(20),
    road_side_code       smallint,
    road_side_name       varchar(20),
    official_land_price  numeric(10),
    data_standard_date   timestamp,
    boundary             geometry(Polygon, 4326) not null,
    center_point         geometry(Point, 4326),
    created_at           timestamp default CURRENT_TIMESTAMP,
    updated_at           timestamp default CURRENT_TIMESTAMP
);