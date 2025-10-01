# Realplot

**Realplot** is a simple and lightweight terminal-based tool for rendering realtime bar graphs.\
You can use it to view sensor data like CPU usage or fan speed.

## Instalation

> [!NOTE]
> You need Go version 1.24.6 or greater. 

### Go

```bash
go install github.com/charadev96/realplot/cmd/realplot@latest
```

### Building from Source

```bash
git clone https://github.com/charadev96/realplot.git
go build ./cmd/realplot
```

## Usage

You need to pipe data into `stdin` in order for it to be displayed, each entry separated by a newline.

```bash
script | realplot --min 0 --max 100
```

### Range

You should use the options `--min` and `--max` to set the desired data range, which the data will be clamped to.

### Colors

You can use the options `--color-border` and `--colors-graph` to set border and graph colors respectfully.
Check `--help` for info on how to format colors.

### Border

You can disable the border around the graph with `--no-border`.

## License

Reaplot is licensed under the [MIT License](LICENSE).
