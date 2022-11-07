# Terraform Provider: Aptible for IaaS

To install the provider locally, run:

```bash
make
```

Or, for mac ARM 64 architectures:

```bash
GOOS=darwin GOARCH=arm64 make
```

Linux:

```bash
GOOS=linux GOARCH=amd64 make
```

## Run

Go to an example workspace as below:

```sh
cd ./examples/aws
```

### Logging in

You can get credentials to work with this provider by:

* setting an environment variable with `APTIBLE_TOKEN` before any `terraform` command (ex: `terraform plan`)
* log in with the [Aptible CLI](https://deploy-docs.aptible.com/docs/cli) (`aptible login`)
* setting a `token` stanza in your provider (like below)

```hcl
provider "aptible" {
  host = var.aptible_host
  token = var.token  # <--- SET THIS VALUE
}
```

### Running Terraform Commands

You should now be able to use your terraform commands without interruption

```bash
terraform init
terraform apply
```

## Dev

### Debug with logs

In an effort to make development faster, we can add some overrides to skip
`terraform init` everytime we make a change to our provider code.  [See here
for details](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers).

Create a file `dev.tfrc` in the root of this project

```
provider_installation {
  # Use /home/developer/tmp/terraform-null as an overridden package directory
  # for the hashicorp/null provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # null provider plugin in the given directory.
  dev_overrides {
    "aptible.com/aptible/aptible-iaas" = "/Users/XXX/terraform-provider-aptible-iaas/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {
    "hashicorp/aws" = "hashicorp/aws"
  }
}
```

Then export an environment variable for the override in the shell where you are
running `terraform apply`:

```
export TF_CLI_CONFIG_FILE=/Users/XXX/terraform-provider-aptible-iaas/dev.tfrc
```

`terraform init` will still attempt to install the plugin so `make
local-install` is still necessary but now when running `terraform apply` it
will use the `dev_override` and **not** use the normal package.

#### Making code changes with `dev_overrides`

So now the flow looks like:

```bash
make local-install # only call this once
# cd into workspace
terraform init # only call this once
```

Then for code changes:

```bash
# make code change
make build-dev
# cd into workspace (or have separate shell open in correct dir)
TF_LOG=DEBUG terraform apply

# make code change
make build-dev
# cd into workspace (or have separate shell open in correct dir)
TF_LOG=DEBUG terraform apply
```

We now have a slightly faster development cycle!

### Debug with delve

Instead of using log-based debugging (via `tflog` package) we can use
debug-based debuging where we have a debugger attached to our provider binary
where we can inspect variables and code execution.

[Read about it more from the official
documentation](https://www.terraform.io/plugin/debugging#debugger-based-debugging)

The following documentation assumes you're using the delve cli.  If you want to
use visual studio code, [read the official
documentation.](https://www.terraform.io/plugin/debugging#visual-studio-code)

The primary difference here is that instead of terraform running the provider
binary, we use `delve` to run the binary.  This has implications for
environment variables.  Instead of exporting `APTIBLE_TOKEN` in the shell where
`terraform apply` is run, you set that environment variable in the shell where
you are running `make debug`.

```bash
export APTIBLE_TOKEN="XXX"
make debug # start dlv headless server
```

This starts delve in headless mode.  The output of running this command will
look something like this:

```bash
$ make debug

API server listening at: 0.0.0.0:33000
# other outputs
Provider started. To attach Terraform CLI, set the TF_REATTACH_PROVIDERS environment variable with the following:

        TF_REATTACH_PROVIDERS='{"aptible.com/aptible/aptible-iaas":{"Protocol":"grpc","ProtocolVersion":6,"Pid":81791,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/tr/86hx1jnj6f19nd3311mgs8340000gn/T/plugin3699290919"}}}'
```

```bash
make dc # attach client to dlv server
```

Set any breakpoints you want:

```bash
b internal/provider/provider.go:123
```

Then inside `dlv` repl continue (`c`) the execution.

Now, copy and export environment variable `TF_REATTACH_PROVIDERS` in the shell where you are
running `terraform apply` and then run the command:

```bash
export TF_LOG=DEBUG # for tflogs
export TF_REATTACH_PROVIDERS='{"aptible.com/aptible/aptible-iaas":{"Protocol":"grpc","ProtocolVersion":6,"Pid":81791,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/tr/86hx1jnj6f19nd3311mgs8340000gn/T/plugin3699290919"}}}'
terraform apply # run the provider
```

You should now see your breakpoints getting activated in your `dlv` repl where
you can inspect variables and step through code execution!

#### Making code changes with delve

After making a code change, type `rebuild` in the `dlv` repl.  This will rebuild
the provider binary and kill `terraform apply`.  Unfortunately because it's a
new process, you'll have to press `c` to continue execution and then copy/paste
the `TF_REATTACH_PROVIDERS` environment variable and then run `terraform apply`
again.  Not ideal but still a pretty speedy dev workflow.
