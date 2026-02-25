## related information

ResourceQuota and additional-guest-memory-overhead-ratio history:

(1) The RQ auto-scaling when VM is migrating, bug and fix

Issue (v121):
https://github.com/harvester/harvester/issues/7178
https://github.com/harvester/harvester/issues/7585

JIRA:
https://jira.suse.com/browse/SURE-9502

Analysis:
https://github.com/w13915984028/harvester-develop-summary/blob/main/v121-vm-migration-with-resources-quota.md#issue-resourcequota-was-changed-unexpected

PR:
https://github.com/harvester/harvester/pull/7504


(2) VM OOM and additional-guest-memory-overhead-ratio (v1.4, v1.5)

Issue:

https://github.com/harvester/harvester/issues/2419
https://github.com/harvester/harvester/issues/2522
https://github.com/harvester/harvester/issues/5768

JIRA:

https://jira.suse.com/browse/SURE-8365
https://jira.suse.com/browse/SURE-8479


PR:

https://github.com/harvester/harvester/pull/6438
https://github.com/harvester/docs/pull/615

Analysis:
https://github.com/w13915984028/harvester-develop-summary/blob/main/oom-related-issue-investigation.md
https://github.com/w13915984028/harvester-develop-summary/blob/main/config-additionalGuestMemoryOverheadRatio-on-harvester-v132.md

the limitation was mentioned https://github.com/w13915984028/harvester-develop-summary/blob/main/config-additionalGuestMemoryOverheadRatio-on-harvester-v132.md#resource-quota



(3) RQ auto-scaling when VM is migrating and the additional-guest-memory-overhead-ratio was also changed (v180)

JIRA:

https://jira.suse.com/browse/SURE-10800

Issue:
https://github.com/harvester/harvester/issues/9753

Analysis:
https://github.com/w13915984028/harvester-develop-summary/blob/main/ts-ref/v170-vm-migration-with-resource-quota-and-overhead-ratio.md

PR:
https://github.com/harvester/harvester/pull/9942


