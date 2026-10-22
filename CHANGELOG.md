# Changelog

## [1.8.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.7.0...v1.8.0) (2026-10-22)


### Features

* **docs:** add data source usage examples and clean template fallbacks ([#42](https://github.com/rixlhq/terraform-provider-rixl/issues/42)) ([8608e09](https://github.com/rixlhq/terraform-provider-rixl/commit/8608e096c8568e9fa993fe23f562f5275c3fdfd5))
* **docs:** add registry guides, resource subcategories, and search engine optimization ([#35](https://github.com/rixlhq/terraform-provider-rixl/issues/35)) ([f31123c](https://github.com/rixlhq/terraform-provider-rixl/commit/f31123c5cad9ee3216c94c9c5b121e6da016ddff))
* **docs:** expand guides, optimize README and enhance SEO keywords ([#39](https://github.com/rixlhq/terraform-provider-rixl/issues/39)) ([dec499f](https://github.com/rixlhq/terraform-provider-rixl/commit/dec499fdf32667693be3459e27fbf9fcc0908102))
* **provider:** bump sdk to v0.9.0 and implement import support with drift detection ([#43](https://github.com/rixlhq/terraform-provider-rixl/issues/43)) ([4c6af41](https://github.com/rixlhq/terraform-provider-rixl/commit/4c6af4165ff41bbfb237e0db1c4be3a806987834))
* **provider:** embed data source descriptions and expand test coverage ([#38](https://github.com/rixlhq/terraform-provider-rixl/issues/38)) ([25aae05](https://github.com/rixlhq/terraform-provider-rixl/commit/25aae05bed8d6b31c7d8298d84b7b250991c2f26))


### Bug Fixes

* add MIT license and update repository licensing ([df02d38](https://github.com/rixlhq/terraform-provider-rixl/commit/df02d388fdf8a45af99f7f8f19b6dd7ee4c934d9))
* **ci:** coordinate goreleaser draft and release-please tag configuration ([4a045e9](https://github.com/rixlhq/terraform-provider-rixl/commit/4a045e9ec95f31177feaf3ca4da4eaf139e9d388))
* **docs:** align custom domain documentation with API and purge CDN references ([#41](https://github.com/rixlhq/terraform-provider-rixl/issues/41)) ([0fa7bd9](https://github.com/rixlhq/terraform-provider-rixl/commit/0fa7bd9ff417fba07baeb406ffa3e4865198f035))
* **docs:** synchronize guide code examples, resource schemas, and cross-references ([#40](https://github.com/rixlhq/terraform-provider-rixl/issues/40)) ([f30b289](https://github.com/rixlhq/terraform-provider-rixl/commit/f30b28929d98a50bf81176f19ddb3238df3c702e))

## [1.7.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.6.0...v1.7.0) (2026-10-19)


### Features

* **docs:** enhance registry documentation, add resource examples, and optimize SEO ([#33](https://github.com/rixlhq/terraform-provider-rixl/issues/33)) ([00bd7cf](https://github.com/rixlhq/terraform-provider-rixl/commit/00bd7cf9c60ee029a29c794006bc4bc4be70ae44))

## [1.6.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.5.0...v1.6.0) (2026-10-19)


### Features

* **provider:** replace legacy generated resource schemas with typed definitions and implement updates ([#31](https://github.com/rixlhq/terraform-provider-rixl/issues/31)) ([cd714f1](https://github.com/rixlhq/terraform-provider-rixl/commit/cd714f1121c988dae7f0f0348cb1d3f7ec3bbbf8))

## [1.5.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.4.0...v1.5.0) (2026-10-19)


### Features

* **provider:** implement strongly typed terraform plugin framework resources ([#29](https://github.com/rixlhq/terraform-provider-rixl/issues/29)) ([58f6276](https://github.com/rixlhq/terraform-provider-rixl/commit/58f62760989590b2b1c9accb26d9618517ab5820))

## [1.4.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.3.1...v1.4.0) (2026-10-19)


### Features

* **provider:** eliminate custom generator scripts and adopt idiomatic Go provider toolchain ([#27](https://github.com/rixlhq/terraform-provider-rixl/issues/27)) ([fb6e8ac](https://github.com/rixlhq/terraform-provider-rixl/commit/fb6e8ac2ac91a8de70fb73ab6da28565e77518aa))

## [1.3.1](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.3.0...v1.3.1) (2026-10-19)


### Bug Fixes

* **generator:** clean timestamp docs and integrate codegen tools into mise ([#24](https://github.com/rixlhq/terraform-provider-rixl/issues/24)) ([a3e7a08](https://github.com/rixlhq/terraform-provider-rixl/commit/a3e7a08cabe5163b8284a559f795e8bdc02b8ce6))

## [1.3.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.2.0...v1.3.0) (2026-10-19)


### Features

* **provider:** sync schemas with latest openapi spec and fix generator pipeline ([#22](https://github.com/rixlhq/terraform-provider-rixl/issues/22)) ([e297504](https://github.com/rixlhq/terraform-provider-rixl/commit/e297504093a8dd66e5b06e4b22289294d2f942a5))

## [1.2.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.1.0...v1.2.0) (2026-10-19)


### Features

* **provider:** wire version into framework metadata and user-agent ([#20](https://github.com/rixlhq/terraform-provider-rixl/issues/20)) ([3d6d006](https://github.com/rixlhq/terraform-provider-rixl/commit/3d6d006ab956ae368352b15da8d6d2c0387f1c78))

## [1.1.0](https://github.com/rixlhq/terraform-provider-rixl/compare/v1.0.0...v1.1.0) (2026-10-19)


### Features

* **provider:** add rixl_auth_provider_connection resource ([89f6432](https://github.com/rixlhq/terraform-provider-rixl/commit/89f64320456df21cffad454d8dbc1d3373c94235))
* **provider:** add rixl_video_chapters collection resource ([84eaf47](https://github.com/rixlhq/terraform-provider-rixl/commit/84eaf47a36b92bf808d453f8ac8a39d0b1973cd5))


### Bug Fixes

* **provider:** add provider schema description for registry docs ([39f0d34](https://github.com/rixlhq/terraform-provider-rixl/commit/39f0d34e9cbd2a2c46b0b7a828770ce67a31903a))
* **provider:** add state upgrade for payment_method provider rename ([f617513](https://github.com/rixlhq/terraform-provider-rixl/commit/f6175137f1952a6e307274eff69e99a1b90316ae))
* **provider:** regenerate schemas with concise RFC 3339 timestamp docs ([ebaca70](https://github.com/rixlhq/terraform-provider-rixl/commit/ebaca700dfc1071f20baac546c99a436ed7a16e8))
* **provider:** rename payment_method provider attr to provider_name ([ddda6e3](https://github.com/rixlhq/terraform-provider-rixl/commit/ddda6e365f4bd8c024626ec4e69520147cdecd8b))
* **provider:** validate chapter uniqueness and ordering, preserve fractional positions on delete, and compensate create failures ([5461f97](https://github.com/rixlhq/terraform-provider-rixl/commit/5461f975c0a3592b226ab7fbf276a33139744076))
* **tools:** include organization and blog_subscription in MANUAL_RESOURCES ([5bf2a46](https://github.com/rixlhq/terraform-provider-rixl/commit/5bf2a46935bf1bacbf56c1f44420e73deee7e03d))
* **tools:** sanitize protobuf Timestamp descriptions in generated spec ([fc2e00c](https://github.com/rixlhq/terraform-provider-rixl/commit/fc2e00caad54a2e46c9235019a326de7badf4092))

## 1.0.0 (2026-10-10)


### Features

* bootstrap Rixl Terraform provider repository ([9cdde05](https://github.com/rixlhq/terraform-provider-rixl/commit/9cdde059f32ecd72dce210c4987072299c81f549))
* **datasource:** map pagination, sort, and filter config to query parameters ([116b2fd](https://github.com/rixlhq/terraform-provider-rixl/commit/116b2fd56e59a22f6165e13d059c6f43a0f2016b))
* expand data source coverage from latest OpenAPI spec ([e7df606](https://github.com/rixlhq/terraform-provider-rixl/commit/e7df606adb537eff05aa6d1410e42cde9da4d428))
* expand data source coverage from latest OpenAPI spec ([fab1292](https://github.com/rixlhq/terraform-provider-rixl/commit/fab129205f0e665fde639a7cfb0fb961bc3c525b))
* implement provider scaffolding, API client, value conversion, feed resource, and data sources ([e100c09](https://github.com/rixlhq/terraform-provider-rixl/commit/e100c0988b2fafdce9b9b7064b434c13b38489ea))
