New package `sperr`

# `sperr.Error`

`sperr.Error` satisfies the standard `error` interface. An `sperr.Error` is a stateful object with a StackTrace of the call stack to a depth of `32` (`32` picked OTA)

## Methods

### `WithMessage(format string, args ...interface())`

Sets the formatted string to `Error` if the `message` property is `empty`. Otherwise creates a new `Error` by wrapping around this `Error` and sets the message on the `wrapper`.

### `WithDetail(format string, args ...interface())`

Sets the formatted string to `Error` if the `detail` property is `empty`. Otherwise creates a new `Error` by wrapping around this `Error` and sets the detail on the `wrapper`.

### `AsRootMessage()`

Sets this `Error` as the root error - effectively hiding all wrapped `errors` under this `error` when an `Error() string` is constructed.

## Create `sperr.Error`:

### `sperr.New(format string, args interface{}...)`

This is to be used when we want to create new `Error` instances. Always carries a `StackTrace`. It is recommended that this function be called from the actual place of the error and not to create error.

### `sperr.Wrap(err error)`

If the given `err` is not an `sperr.Error`, this wraps around `err` and creates an `sperr.Error` along with a `StackTrace`. Returns `nil` if `err` is `nil`. `Wrap` tries to infer a friendly message for the error and if the inference succeeded, it will set the friendly message as it's own message.

### `sperr.Wrapf(err error, format string, args ...interface{})`

Wrap an `error` to create an `sperr.Error` and sets a formatted message to the `wrapper`. `Wrapf` is functionally equivalent to `Wrap(err).WithMessage()` - but maintains the proper call stack.

## Examples:

### Create a new `error`

```
dbState, err := GetState()
if err != nil {
  log.Println("[TRACE] Error while loading database state", err)
  return err
}
if dbState != nil {
  return sperr.New("cannot install db - a previous version of the Steampipe service is still running. To stop running services, use %s ", constants.Bold("steampipe service stop"))
}
```

### Wrap an `error`

```
if err := json.Unmarshal(bytContent, &data); err != nil {
  return nil, sperr.Wrap(err)
}
```

### Wrap an `error` with a `message`

```
if err := json.Unmarshal(byteContent, &data); err != nil {
  return nil, sperr.Wrapf(err, "error unmarshalling file content in %s", filePath)
}
```

or

```
if err := json.Unmarshal(byteContent, &data); err != nil {
  return nil, sperr.Wrap(err).WithMessage("error unmarshalling file content in %s", filePath)
}
```

### Convert `any` value to an `Error`

```
defer func() {
  if r := recover(); r != nil {
    err = sperr.ToError(r)
  }
}()
```

### Create an `error` with `message` and `detail`

```
func validateData(data int) error {
  if data > 10 {
    return sperr.New("invalid argument: %d", data).WithDetail("error occurred with %d argument", data)
  }
  return nil
}
```

### Wrap an `error` with `detail`

```
err := validateData(userInput.numAttacks)
if err!= nil {
  return sperr.Wrap(err).WithDetail("error occurred with %d argument", userInput.numAttacks)
}
```

### Wrap an `error` with `message`

```
err := validateData(userInput.numAttacks)
if err!= nil {
  return sperr.Wrap(err).WithMessage("error occurred with %d argument", userInput.numAttacks)
}
```

> While wrapping around `err`, if `Wrap` could infer a `message` then `WithMessage` will create a `wrapper` around the **output of `sperr.Wrap(err)`** and set the message on the `wrapper`. Otherwise, it will just set the `message` on `err`.

### Wrap an `error` with formatted message and then set a `message`

```
err = sperr.Wrapf(err, "error occurred").WithMessage("error occurred with %d argument", intArgument)
```

### Set an `error` as the `root` error

```
err = sperr.Wrap(err).SetRoot()
```

> Setting an error as the `root` error hides all errors below it from the user interface. They are not purged - just hidden from display when displaying error messages. When enumerating error `details`, the details of all errors in the stack are shown - including errors under a `root` error.

## Technicalities

### Wrapping as necessary

#### `Wrap`

The package function `Wrap` wraps around a given `error` instance if and only if it is not an instance of `sperr.Error`. This effectively ensures that the return of `Wrap` is always an instance of `sperr.Error`.

#### `Wrapf`

The package function `Wrapf` **always** wraps around the `error` given to it. This is because `Wrapf` always sets it's own message with the arguments provided.

#### `WithMessage`

`WithMessage` sets the internal `message` if it is empty. Otherwise, it will create a `wrapper` around it's instance and set the `message` on the `wrapper` and returns the `wrapper`. This ensures that `WithMessage` is never lossy - but only creates wrappers when necessary.

#### `WithDetail`

`WithDetail` behaves just like `WithMessage`, but on the `detail` property.

#### Example:

> ```
> err = sperr.Wrapf(err, "error occurred").
>              WithDetail("added detail").
>              WithMessage("error occurred with %d argument", intArgument)
> ```
>
> Result:
>
> ```
> Error {
>   Error {
>     err
>     Message : "error occurred"
>     Detail  : "added detail"
>   }
>   Message : "error occurred with %d argument"
> }
> ```
