# DB Hose

![DB Hose Logo](https://dbhose.com/logo.png)

DB Hose is an open-source database migration management tool designed for enterprise developers. It provides secure storage, easy integration, and comprehensive logging for database migration across various environments.

[![GitHub stars](https://img.shields.io/github/stars/dbhose/dbhose.svg)](https://github.com/IamNator/db-hose.com/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Pulls](https://img.shields.io/docker/pulls/dbhose/dbhose.svg)](https://hub.docker.com/r/dbhose/dbhose)

## Features

- **Secure Credential Storage**: AES-256 encryption for all stored credentials
- **Flexible Integration**: Easy to integrate with popular database systems and DevOps tools
- **Audit Logging**: Comprehensive logging of all credential access and modifications
- **Role-Based Access Control**: Granular control over who can access and modify credentials
- **API-First Design**: RESTful API for seamless integration with your existing tools and workflows
- **Self-Hosted or Managed**: Choose between self-hosted deployment or our managed service

## Quick Start

### Using Docker

```bash
docker pull dbhose/dbhose:latest
docker run -d -p 8080:8080 dbhose/dbhose:latest
```

### Manual Installation

1. Ensure you have Go 1.16+ installed
2. Clone the repository:
   ```bash
   git clone https://github.com/dbhose/dbhose.git
   cd dbhose
   ```
3. Build the project:
   ```bash
   go build
   ```
4. Run DB Hose:
   ```bash
   ./dbhose
   ```

Visit `http://localhost:8080` to access the DB Hose dashboard.


## Deployment Options

### Self-Hosted

To deploy DB Hose on your own infrastructure, follow our [self-hosted deployment guide](https://docs.dbhose.com/self-hosted).

### Managed Service

For information about our managed service offering, visit [https://db-hose-com.onrender.com/](https://db-hose-com.onrender.com/).

## Contributing

We welcome contributions to DB Hose! Please see our [contributing guidelines](CONTRIBUTING.md) for more information on how to get started.

## Security

If you discover a security vulnerability within DB Hose, please send an e-mail to security@dbhose.com. All security vulnerabilities will be promptly addressed.

## License

DB Hose is open-source software licensed under the [MIT license](LICENSE.md).

## Support

- Documentation: [https://docs.dbhose.com](https://docs.dbhose.com)
- Issue Tracker: [https://github.com/dbhose/dbhose/issues](https://github.com/dbhose/dbhose/issues)
- Community Forum: [https://community.dbhose.com](https://community.dbhose.com)

## Stay in Touch

- Twitter: [@DBHose](https://twitter.com/DBHose)
- Blog: [https://dbhose.com/blog](https://dbhose.com/blog)

---

DB Hose - Secure Database Migration Management for Enterprise Developers
