# connection and plugin config tests

## 1 Connection has invalid plugin 

```hcl
connection "aws" {
  plugin = "aws_bar"
}
```

### Expected

#### On interactive startup: 
```
Warning: 1 plugin required by connection is missing. To install, please run steampipe plugin install aws_bar1
```

#### On file watcher event:
```
Warning: 1 plugin required by connection is missing. To install, please run steampipe plugin install aws_bar1
```

### Actual

As expected

## 2 Startup with invalid plugin (referring to instance by name) 

```hcl
connection "aws" {
  plugin = "aws_bar"
}


plugin "aws_bar"{
  source="aws"
}
```

Expected and actual as 1

## 3 Connection referring to valid plugin instance, and plugin instance referring to invalid plugin

```hcl
connection "aws" {
  plugin = plugin.aws_bar
}


plugin "aws_bar"{
  source="aws_bad"
}
```

### Expected


#### On interactive startup:
```
Warning: 1 plugin required by connection is missing. To install, please run steampipe plugin install aws_bar1
```

#### On file watcher event startup:
```
Warning: 1 plugin required by connection is missing. To install, please run steampipe plugin install aws_bar1
```

### Actual


#### On interactive startup:

RefreshConnections stalls

#### On file watcher event:

nothing happens?

## 4 Connection referring to invalid plugin instance

```hcl
connection "aws" {
  plugin = plugin.aws_bar
}
```

### Expected


#### On interactive startup:
```
Warning: counld not resolve plugin
```

#### On file watcher event startup:
```
Warning: counld not resolve plugin
```

### Actual


#### On interactive startup:

Connection not loaded, no error

#### On file watcher event:

Connection not loaded, no error