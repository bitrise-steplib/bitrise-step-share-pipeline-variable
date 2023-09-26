### Example

```yaml
steps:
- script@1:
    title: Should we run UI tests?
    inputs:
    - content: |-
        set -eo pipefail
        # Custom logic goes here
        envman add --key RUN_UI_TESTS --value true
- share-pipeline-variable@1:
    title: Configure next pipeline stage
    inputs:
    - variables: |-
        RUN_UI_TESTS
        BUILD_TYPE=debug
```
