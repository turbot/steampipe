## Description

This control checks whether the following logs of Amazon RDS are enabled and sent to CloudWatch Logs:

- Oracle: (Alert, Audit, Trace, Listener)
- PostgreSQL: (Postgresql, Upgrade)
- MySQL: (Audit, Error, General, SlowQuery)
- MariaDB: (Audit, Error, General, SlowQuery)
- SQL Server: (Error, Agent)
- Aurora: (Audit, Error, General, SlowQuery)
- Aurora-MySQL: (Audit, Error, General, SlowQuery)
- Aurora-PostgreSQL: (Postgresql, Upgrade).

RDS databases should have relevant logs enabled. Database logging provides detailed records of requests made to RDS. Database logs can assist with security and access audits and can help to diagnose availability issues.

## Remediation

Logging options are contained in the DB parameter group associated with the RDS DB cluster or instance. To enable logging when the default parameter group for the database engine is used, you must create a new DB parameter group that has the required parameter values. You must then associate the customer DB parameter group with the DB cluster or instance.

To enable and publish MariaDB, MySQL, or PostgreSQL logs to CloudWatch Logs from the AWS Management Console, set the following parameters in a custom DB Parameter Group:

**MariaDB**

  - `general_log` = **1**
  - `slow_query_log` = **1**
  - `log_output` = **FILE**

**MySQL**

  - `general_log` = **1**
  - `slow_query_log` = **1**
  - `log_output` = **FILE**

**PostgreSQL**

  - `log_statement` = **all**
  - `log_min_duration_statement` = *minimum query duration (ms) to log*

**To create a custom DB parameter group**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Parameter groups`.
3. Choose `Create parameter group`. The Create parameter group window appears.
4. In the Parameter group family list, choose a DB parameter group family.
5. In the `Type` list, choose `DB Parameter Group`.
6. In `Group name`, enter the name of the new DB parameter group.
7. In `Description`, enter a description for the new DB parameter group.
8. Choose `Create`.

**To create a new option group for MariaDB logging by using the console**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Option groups`.
3. Choose `Create group`.
4. In the `Create option group` window, do the following:
   - For Name, type a name for the option group that is unique within your AWS account. The name can contain only letters, digits, and hyphens.
   - For Description, type a brief description of the option group. The description is used for display purposes.
   - For Engine, choose the DB engine that you want.
   - For Major engine version, choose the major version of the DB engine that you want.
5. To continue, choose `Create`.
6. Choose the name of the option group you just created.
7. Choose `Add option`.
8. Choose `MARIADB_AUDIT_PLUGIN` from the `Option name` list.
9. Set `SERVER_AUDIT_EVENTS` to `CONNECT`, `QUERY`, `TABLE`, `QUERY_DDL`, `QUERY_DML`, `QUERY_DCL`.
10. Choose `Add option`.

**To publish SQL Server DB, Oracle DB, or PostgreSQL logs to CloudWatch Logs from the AWS Management Console**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Databases`.
3. Choose the DB instance that you want to modify.
4. Choose `Modify`.
5. Under `Log exports`, choose all of the log files to start publishing to CloudWatch Logs.
6. `Log exports` is available only for database engine versions that support publishing to CloudWatch Logs.
7. Choose `Continue`. Then on the summary page, choose `Modify DB Instance`.

**To apply a new DB parameter group or DB options group to an RDS DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Databases`.
3. Choose the DB instance that you want to modify.
4. Choose `Modify`. The `Modify DB Instanc`e page appears.
5. Under `Database options`, change the DB parameter group and DB options group as needed.
6. When you finish you changes, choose `Continue`. Check the summary of modifications.
7. (Optional) Choose `Apply immediately` to apply the changes immediately. Choosing this option can cause an outage in some cases. For more information, see Using the [Apply Immediately setting](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html#USER_ModifyInstance.ApplyImmediately).
8. Choose `Modify DB Instance` to save your changes.