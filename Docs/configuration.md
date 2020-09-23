# Configuration
Tensorflow Deploy allows to use one of three different configuration sources: CLI, environments or YAML file. The priority of sources is like: CLI => ENV => FILE(YAML) what means:
- FILE (YAML) has the lowest priority
- ENV has higher priority than YAML file
- CLI has higher priority than ENV

## Documentation
Documentation of possible configurations:

*  [ENV Parameters](configuration-envs.md)
*  [CLI Parameters](configuration-cli.md)
*  [YAML File Parameters](configuration-yaml.md)
