#!/bin/sh

# Set default UID and GID (1000) if not provided
UID=${UID:-1000}
GID=${GID:-1000}

# Change ownership of the directory
chown -R ${UID}:${GID} /pastes

# Execute the main as the specified user
exec su-exec ${UID}:${GID} ./main "$@"
