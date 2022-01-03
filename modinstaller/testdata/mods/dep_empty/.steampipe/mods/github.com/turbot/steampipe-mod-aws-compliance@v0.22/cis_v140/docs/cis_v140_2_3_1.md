## Description

Amazon RDS encrypted DB instances use the industry standard AES-256 encryption algorithm to encrypt your data on the server that hosts your Amazon RDS DB instances. After your data is encrypted, Amazon RDS handles authentication of access and decryption of your data transparently with a minimal impact on performance.

Databases that hold sensitive and critical data, it is highly recommended to implement encryption in order to protect your data from unauthorized access. With RDS encryption enabled, the data stored on the instance underlying storage, the automated backups, Read Replicas, and snapshots, become all encrypted.

## Remediation

### From Console

1. Login to the AWS [RDS console](https://console.aws.amazon.com/rds/).
2. In the left navigation panel, click on `Databases`
3. Select the Database instance that needs to encrypt.
4. Click on **Actions** button placed at the top right and select **Take Snapshot**.
5. On the Take Snapshot page, enter a database name of which want to take snapshot in the Snapshot Name field and click Take Snapshot.
6. Select the newly created snapshot and click the **Copy** from the dashboard top menu.
7. On the Make Copy of DB Snapshot page, perform the following:
   1. In the New DB Snapshot Identifier field, Enter a name for the `new snapshot`.
   2. Check `Copy Tags`, New snapshot must have the same tags as the source snapshot.
   3. Select Yes from the **Enable Encryption** dropdown list to enable encryption, Can choose to use the AWS default encryption key or custom key from Master Key dropdown list.
8. Click **Copy Snapshot** to create an encrypted copy of selected instance snapshot.
9. Select the new Snapshot Encrypted Copy and click Restore Snapshot button from the dashboard top menu, This will restore the encrypted snapshot to a new database instance.
10. On the Restore DB Instance page, enter a unique name for the new database instance in the DB Instance Identifier field.
11. Review the instance configuration details and click **Restore DB Instance**.
12. As the new instance provisioning process is completed can update application configuration to refer to the endpoint of the new Encrypted database instance once the database endpoint is changed at the application level, can remove the unencrypted instance.

### From Command Line

1. Run describe-db-instances command to list all RDS database names available in the selected AWS region, The command output should return database instance identifier.

```bash
aws rds describe-db-instances --region <region-name> --query 'DBInstances[*].DBInstanceIdentifier'
```

2. Run create-db-snapshot command to create a snapshot for the selected database instance, The command output will return the new snapshot with name DB
Snapshot Name.

```bash
aws rds create-db-snapshot --region <region-name> --db-snapshot-identifier <DB-Snapshot-Name> --db-instance-identifier <DB-Name>
```

3. Now run list-aliases command to list the KMS keys aliases available in a specified region, The command output should return each key alias currently available. For our RDS encryption the activation process, locate the ID of the AWS default KMS key.

```bash
aws kms list-aliases --region <region-name>
```

4. Run copy-db-snapshot command using the default KMS key ID for RDS instances returned earlier to create an encrypted copy of the database instance snapshot, the command output will return the encrypted instance snapshot configuration.

```bash
aws rds copy-db-snapshot --region <region-name> --source-db-snapshotidentifier <DB-Snapshot-Name> --target-db-snapshot-identifier <DB-SnapshotName-Encrypted> --copy-tags --kms-key-id <KMS-ID-For-RDS>
```

5. Run restore-db-instance-from-db-snapshot command to restore the encrypted snapshot created at the previous step to a new database instance, if successful, the command output should return the new encrypted database instance configuration.

```bash
aws rds restore-db-instance-from-db-snapshot --region <region-name> --dbinstance-identifier <DB-Name-Encrypted> --db-snapshot-identifier <DBSnapshot-Name-Encrypted>
```

6. Run describe-db-instances command to list all RDS database names, available in the selected AWS region, output will return database instance identifier name. Select encrypted database name that we just created DB-Name-Encrypted.

```bash
aws rds describe-db-instances --region <region-name> --query 'DBInstances[*].DBInstanceIdentifier'
```

7. Run again describe-db-instances command using the RDS instance identifier returned earlier, to determine if the selected database instance is encrypted, the command output should return the encryption status True.

```bash
aws rds describe-db-instances --region <region-name> --db-instance-identifier <DB-Name-Encrypted> --query 'DBInstances[*].StorageEncrypted'
```
