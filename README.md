# backstagectl

A CLI written in Go to interact with a Backstage IDP instance.

## Installation

To install `backstagectl`, follow these steps:

1. **Clone the repository:**

   ```bash
   git clone https://github.com/vcaldaralo/backstagectl.git
   cd backstagectl
   ```

2. **Build the CLI:**
   Make sure you have Go installed on your machine. Then run:

   ```bash
   go build -o backstagectl
   ```

3. **Move the binary to your PATH:**
   You can move the `backstagectl` binary to a directory in your PATH for easier access:

   ```bash
   sudo mv backstagectl /usr/local/bin/
   ```

## Usage

To use BackstageCTL, run the following command in your terminal:

```bash
backstagectl [command] [options]
```

### Commands

- `auth`: Manage authentication with the Backstage IDP.
- `get`: Display one or many Backstage entities
- `check`: Check properties of Backstage entities

### Example

To authenticate with Backstage, use the following command:

```bash
backstagectl auth --baseUrl BACKSTAGE_URL  --token YOUR_TOKEN
```

or

```bash
backstagectl auth \
   --baseUrl BACKSTAGE_URL  \
   --tls-cert YOUR_TLS_CERT_PATH \
   --tls-key YOUR_TLS_KEY_PATH
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
