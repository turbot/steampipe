# Steampipe data files

## .steampipe/db

- `versions.json` - Stores information about the embedded database and the FDW installed. Contains information like image_digest, installed_from, version etc. Removing this file would result in losing your database information, and running steampipe would re-install the database and the FDW and hence re-create the file with the latest information.

## .steampipe/internal

- `.passwd` - Stores the database password. Deleting the file does not effect steampipe, you can view your password by using the --show-password flag along with the service commands. Starting the service would re-create the file.

- `pipes.turbot.com.sptt` - Stores the [Turbot Pipes](https://pipes.turbot.com) token. Deleting the file would require you to run steampipe login again.

- `connection.json` - Stores the connection config information. This file gets re-generated everytime RefreshConnections is called.

- `history.json` - Stores the last used queries. Deleting this file would result in losing your history of queries. This file gets re-generated.

- `plugin_manager.json` - Stores plugin manager related information. This file gets created when service is running, and also gets deleted when the service is stopped.

- `steampipe.json` - Stores steampipe service related information. This file gets created when service is running, and also gets deleted when the service is stopped.

- `update_check.json` - Stores the installation state(last_checked and installation_id). Deleting the file would run the update check and re-create the file.

## .steampipe/plugins

- `versions.json` - Stores information about all the plugins installed. Contains information like version, image_digest, binary_digest, binary_arch, installedFrom etc. Removing this file would result in losing your plugin information(incorrect version), and you would need to re-install all your plugins.
``