# Share env vars between stages

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/bitrise-step-share-pipeline-variable?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/bitrise-step-share-pipeline-variable/releases)

Share environment variables between pipeline stages

<details>
<summary>Description</summary>

Share environment variables between pipeline stages
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/steps-and-workflows/steps-and-workflows-index/).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `variables` | A newline (`\n`) separated list of key - value pairs (`{key}={value}`).  The input uses a `{key}={value}` syntax. The first equals sign (`=`) is the delimiter between the key and value of the environment variable. A shorthand syntax of `ENV_KEY` can be used for `ENV_KEY=$ENV_KEY` when sharing an existing environment variable (ENV_KEY).  Examples: ``` MY_ENV_KEY=my value EXISTING_ENV_KEY ``` | required |  |
| `app_url` | The app's URL on Bitrise.io. | required | `$BITRISE_APP_URL` |
| `build_slug` | The build's slug on Bitrise.io. | required | `$BITRISE_BUILD_SLUG` |
| `build_api_token` | API Token for the build on Bitrise.io. | required, sensitive | `$BITRISE_BUILD_API_TOKEN` |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/bitrise-step-share-pipeline-variable/pulls) and [issues](https://github.com/bitrise-steplib/bitrise-step-share-pipeline-variable/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)
