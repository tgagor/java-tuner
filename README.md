# java-tuner

A small, statically linked CLI tool that simplifies tuning of Java applications for containerized environments. It automatically detects available CPU and memory resources, then generates optimal JVM options to improve performance and reliability.

## What does it do?

`java-tuner` inspects your container or host environment to determine CPU and memory limits, then outputs recommended JVM flags. These flags help Java applications run efficiently by:

- Setting memory usage based on available RAM and a configurable percentage.
- Configuring JVM to use the correct number of CPUs.
- Applying sensible defaults for server-class JVM, DNS caching, string deduplication, and more.

This automation removes the guesswork from JVM tuning, especially in dynamic or resource-constrained environments like Docker containers and Kubernetes.

## Why tune Java in containers?

Java's default resource detection often fails in containers, leading to poor performance or crashes. By explicitly setting JVM options based on actual limits, you ensure:

- Predictable memory usage (avoiding OOM errors)
- Efficient CPU utilization
- Faster startup and better runtime stability

## Usage

Run the CLI to print recommended JVM options:

```sh
java-tuner
```

You can set environment variables to override detection:

- `JAVA_TUNER_CPU` — Force CPU count
- `JAVA_TUNER_MEM_PERCENTAGE` — Set max RAM percentage (default: 80)
- `JAVA_TUNER_OPTS` — Add extra JVM flags

## Flags

- `--dry-run, -d` — Print actions but don't execute them
- `--no-color` — Disable color output
- `--verbose, -v` — Increase verbosity of output (shows debug info)
- `--version, -V` — Display the application version and exit

## Typical use cases

- **Docker Entrypoint**: Use `java-tuner` to generate JVM flags before starting your Java app.
- **Kubernetes**: Ensure your Java app respects pod resource limits.
- **CI/CD**: Validate JVM options for different environments.

## Example

```sh
JAVA_TUNER_CPU=2 JAVA_TUNER_MEM_PERCENTAGE=75 java-tuner
```

Outputs JVM flags like:

```
-XX:+AlwaysActAsServerClassMachine
-XX:ActiveProcessorCount=2
-XX:MaxRAMPercentage=75.0
-XX:MaxRAM=900m
...
```

## How it works

- Reads cgroup files to detect CPU and memory limits.
- Applies defaults and user overrides.
- Outputs JVM flags for use in your Java command.

## License

[GPLv3](./LICENSE)
