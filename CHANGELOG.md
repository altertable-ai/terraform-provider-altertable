# Changelog

## [0.4.0](https://github.com/altertable-ai/terraform-provider-altertable/compare/v0.3.1...v0.4.0) (2026-07-16)


### Features

* add resource identity to environment, service_account, user ([a798ec4](https://github.com/altertable-ai/terraform-provider-altertable/commit/a798ec4bb5f43badfc4bed32e4a57c6e600ca419))
* add user resource ([d4db617](https://github.com/altertable-ai/terraform-provider-altertable/commit/d4db6176d359353d74a8c638d374683fe9c4e1f2))
* add user resource ([eceaf19](https://github.com/altertable-ai/terraform-provider-altertable/commit/eceaf19ab829d3de136d15e84aa6da46b382f670))
* **catalog:** add altertable_catalog resource ([5557361](https://github.com/altertable-ai/terraform-provider-altertable/commit/5557361f2c60b41b606e23b8a51dcde606974d56))
* **catalog:** add database/connection mapping for the catalog facade ([93bec7c](https://github.com/altertable-ai/terraform-provider-altertable/commit/93bec7c3227bb1f1935810538c5d082563213535))
* **catalog:** add resource identity with colon-string back-compat import ([1903745](https://github.com/altertable-ai/terraform-provider-altertable/commit/1903745d75cf1e084d329dbc0569d788c6d3f3c2))
* **catalog:** data source probes databases then connections ([296cb13](https://github.com/altertable-ai/terraform-provider-altertable/commit/296cb13d957665941afc0d7065f9079dbb78dc5b))
* **catalog:** route resource to databases or connections by engine ([d29edfe](https://github.com/altertable-ai/terraform-provider-altertable/commit/d29edfe88681d7e4eb32b72e633fae1d869f8e3e))
* **client:** add HTTP client core and stubbed entity methods ([b53d1a5](https://github.com/altertable-ai/terraform-provider-altertable/commit/b53d1a5b3de9d021e9bbbd5aa2e94d8fda47aff3))
* **client:** implement REST methods for environments, service accounts, connections, databases, credentials ([cbe2f6d](https://github.com/altertable-ai/terraform-provider-altertable/commit/cbe2f6d9dedfb4a4f3a9d59957b28c522a407698))
* **credential:** add altertable_credential resource with create-once secret ([655cf24](https://github.com/altertable-ai/terraform-provider-altertable/commit/655cf246ccf02114b9f4e5d8b5e84e8abe75ba03))
* **credential:** add resource identity with colon-string back-compat import ([9fa1776](https://github.com/altertable-ai/terraform-provider-altertable/commit/9fa1776ec42cd4d9bf95f7c317524fe013e53d31))
* **credential:** data source keyed by principal and environment ([b64a6f6](https://github.com/altertable-ai/terraform-provider-altertable/commit/b64a6f6bbe887d03aedf1cdd935b8681ac48fa91))
* **credential:** discriminated principal resource with write-once password ([fb17f76](https://github.com/altertable-ai/terraform-provider-altertable/commit/fb17f762a8f4601fc7ee0fc250a1526eb02676a5))
* **data-sources:** add environment, catalog, and user lookups ([b881b67](https://github.com/altertable-ai/terraform-provider-altertable/commit/b881b6777ac0cdafac1be51f6ec660b0fecd333e))
* **data-sources:** add service account and credential lookups ([2cc7ba4](https://github.com/altertable-ai/terraform-provider-altertable/commit/2cc7ba4f70ec8434521778b7ae4a413affeabc7b))
* **environment:** add altertable_environment resource ([7b0255d](https://github.com/altertable-ai/terraform-provider-altertable/commit/7b0255d76bcaac4d529f6f42ab97caa483ba649f))
* **environment:** look up data source by slug or id ([e9929a8](https://github.com/altertable-ai/terraform-provider-altertable/commit/e9929a80e8ac542c6b0d6de9da55204914538de0))
* **environment:** wire resource to REST API with immutable attributes ([1ebfc2d](https://github.com/altertable-ai/terraform-provider-altertable/commit/1ebfc2d258edcd88763e74634aeaf437e76613ef))
* **provider:** add provider, entrypoint, config resolution, schema gate ([b35e4a5](https://github.com/altertable-ai/terraform-provider-altertable/commit/b35e4a5b4720a518fe0c3f80967457a2ba1fe526))
* **role-set:** add altertable_role_set resource with principal validator ([718814a](https://github.com/altertable-ai/terraform-provider-altertable/commit/718814a23cb0a2b636ac02f959c750381d4ab3b7))
* **service-account:** add altertable_service_account resource ([7758f94](https://github.com/altertable-ai/terraform-provider-altertable/commit/7758f941eb5955d97e4fffb0fd0138f193048c75))
* **service-account:** look up data source by id ([1c26241](https://github.com/altertable-ai/terraform-provider-altertable/commit/1c262412f13423d7d07cc5d583503e2fc6c333c3))
* **service-account:** wire resource to REST API using label ([de09789](https://github.com/altertable-ai/terraform-provider-altertable/commit/de09789e9f743eb1a01711ec359a0ff8d9172b81))
* **user:** add altertable_user resource ([88d7fdb](https://github.com/altertable-ai/terraform-provider-altertable/commit/88d7fdbb3fdea6f473275883a812d3ce9627c207))
* **users,roles:** wire user lookup + role assignments; drop read-only user resource ([61f2ffb](https://github.com/altertable-ai/terraform-provider-altertable/commit/61f2ffb209f897c01628d5f8dedf3f0d1a21b9ce))
* **whoami:** add whoami data source and fail-fast credential validation ([93aee2a](https://github.com/altertable-ai/terraform-provider-altertable/commit/93aee2a52ce7985ed03565446eea59d4409db924))
* working provider implementation for a first E2E test against local server ([b2a0f3f](https://github.com/altertable-ai/terraform-provider-altertable/commit/b2a0f3f3eca48a71cb539770d448ef704e6d7d9b))


### Bug Fixes

* align github org and provider namespace ([8193c59](https://github.com/altertable-ai/terraform-provider-altertable/commit/8193c59ba230aad55be55691b0169d262c4e52f4))
* align github org and provider namespace ([4c826bc](https://github.com/altertable-ai/terraform-provider-altertable/commit/4c826bc5e03da7ce998620e94d6034eaa35443bb))
* **catalog:** make snapshot_retention_days Optional+Computed to avoid inconsistent apply ([0268b8a](https://github.com/altertable-ai/terraform-provider-altertable/commit/0268b8ac0091d4b620e970564d3612517a5b7692))
* **catalog:** set known db-only fields for connection engines; make tags/description computed ([70f263a](https://github.com/altertable-ai/terraform-provider-altertable/commit/70f263a2d6cff6494b99789d90c4080a91f8705c))
* **ci:** use draft releases for immutable releases support ([2c74f74](https://github.com/altertable-ai/terraform-provider-altertable/commit/2c74f7467f0da2e71606144d5cea3c268ebda633))
* **ci:** use draft releases for immutable releases support ([41f8224](https://github.com/altertable-ai/terraform-provider-altertable/commit/41f822493479bfc55a61e0d16f5ed8879474061d))
* clean repo initialization ([c92322e](https://github.com/altertable-ai/terraform-provider-altertable/commit/c92322ec4fa85f099a6cbfabe3f9197f5c237ea7))
* **credential:** preserve write-once password explicitly in Update ([9091edc](https://github.com/altertable-ai/terraform-provider-altertable/commit/9091edcb49f6b9d6f059a72b5315a4902c8d4751))
* make environment resource provider_region mandatory ([4545311](https://github.com/altertable-ai/terraform-provider-altertable/commit/45453111d262cdaca0230d6a56c01b60af62c2bb))
* make environment resource provider_region mandatory ([8a58add](https://github.com/altertable-ai/terraform-provider-altertable/commit/8a58addde18f664efc1bbfdf1fa86b23e93578ef))
* provider produced inconsistent result after apply because of updated_at ([af2488a](https://github.com/altertable-ai/terraform-provider-altertable/commit/af2488a0da92188518d01ae3b64e574854ca0cd6))
* provider produced inconsistent result after apply because of updated_at ([c83f4b7](https://github.com/altertable-ai/terraform-provider-altertable/commit/c83f4b7f4a67f3fe0345bcae20a4852697d68809))

## [0.3.1](https://github.com/altertable-ai/terraform-provider-altertable/compare/v0.3.0...v0.3.1) (2026-07-16)


### Bug Fixes

* **ci:** use draft releases for immutable releases support ([2c74f74](https://github.com/altertable-ai/terraform-provider-altertable/commit/2c74f7467f0da2e71606144d5cea3c268ebda633))
* **ci:** use draft releases for immutable releases support ([41f8224](https://github.com/altertable-ai/terraform-provider-altertable/commit/41f822493479bfc55a61e0d16f5ed8879474061d))

## [0.3.0](https://github.com/altertable-ai/terraform-provider-altertable/compare/v0.2.0...v0.3.0) (2026-07-16)


### Features

* add user resource ([d4db617](https://github.com/altertable-ai/terraform-provider-altertable/commit/d4db6176d359353d74a8c638d374683fe9c4e1f2))
* add user resource ([eceaf19](https://github.com/altertable-ai/terraform-provider-altertable/commit/eceaf19ab829d3de136d15e84aa6da46b382f670))


### Bug Fixes

* make environment resource provider_region mandatory ([4545311](https://github.com/altertable-ai/terraform-provider-altertable/commit/45453111d262cdaca0230d6a56c01b60af62c2bb))
* make environment resource provider_region mandatory ([8a58add](https://github.com/altertable-ai/terraform-provider-altertable/commit/8a58addde18f664efc1bbfdf1fa86b23e93578ef))
* provider produced inconsistent result after apply because of updated_at ([af2488a](https://github.com/altertable-ai/terraform-provider-altertable/commit/af2488a0da92188518d01ae3b64e574854ca0cd6))
* provider produced inconsistent result after apply because of updated_at ([c83f4b7](https://github.com/altertable-ai/terraform-provider-altertable/commit/c83f4b7f4a67f3fe0345bcae20a4852697d68809))
