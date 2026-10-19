# Changelog

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
