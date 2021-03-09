### Features


### Improvements

- [sdk/go] Add helpers to convert raw Go maps and arrays to Pulumi `Map` and `Array` inputs.
  [#6337](https://github.com/pulumi/pulumi/pull/6337)

- [sdk/go] Return zero values instead of panicing in `Index` and `Elem` methods.
  [#6338](https://github.com/pulumi/pulumi/pull/6338)

- Updating Pulumi to use Go 1.16
  [#6470](https://github.com/pulumi/pulumi/pull/6470)

- [automation/go] - BREAKING - Expose structured logging for Stack.Up/Preview/Refresh/Destroy.
  [#6436](https://github.com/pulumi/pulumi/pull/6436)
  
This change is marked breaking because it changes the shape of the `PreviewResult` struct.

**Before**

```go
type PreviewResult struct {
  Steps         []PreviewStep  `json:"steps"`
  ChangeSummary map[string]int `json:"changeSummary"`
}
```

**After**

```go
type PreviewResult struct {
  StdOut        string
  StdErr        string
  ChangeSummary map[string]int
}
```

### Bug Fixes

