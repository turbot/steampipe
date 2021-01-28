## v0.1.1 [2021-01-28]

_What's new?_
* A new meta-command `.help` has been added.  ([#54](https://github.com/turbot/steampipe/issues/54))
* After `steampipe plugin install`, a link to the plugin docs is displayed.
* A spinner is now displayed for slow queries. ([#77](https://github.com/turbot/steampipe/issues/77))
* A maximum column width of 1024 is now enforced - content longer than this will wrap. ([#12](https://github.com/turbot/steampipe/issues/12))
* The `description` column of the `.inspect` command now fills the available horizontal screen space. ([#11](https://github.com/turbot/steampipe/issues/11))
* The Linux installation package now uses tar instead of zip. ([#63](https://github.com/turbot/steampipe/issues/63))

_Bug fixes_
* Fix results paging failure for very long rows (> 64k chars). ([#75](https://github.com/turbot/steampipe/issues/75))
* Fix invalid query resulting in the database session remaining open. ([#60](https://github.com/turbot/steampipe/issues/60))
* Fix data formatting in json output. ([#14](https://github.com/turbot/steampipe/issues/14))
* Fix incorrect plugin hub link.
* Fix `steampipe query` panic when exiting after `service stopped --force` has been run. ([#38](https://github.com/turbot/steampipe/issues/38))
* Fix `runtime error: slice bounds out of range [1:0]`.  ([#40](https://github.com/turbot/steampipe/issues/40))
* Fix boolean meta-command showing wrong status when no parameter is passed. ([#48](https://github.com/turbot/steampipe/issues/48))
