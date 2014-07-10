#!/bin/sh
{
  C66_CLIENT_URL="https://app.cloud66.com/toolbelt/osx"

  echo "This script requires superuser access to install software."
  echo "You will be prompted for your password by sudo."

  # clear any previous sudo permission
  sudo -k

  # run inside sudo
  sudo sh <<SCRIPT

  # download and extract the client tarball
  rm -rf /usr/local/cloud66
  mkdir -p /usr/local/cloud66
  cd /usr/local/cloud66

  if [[ -z "$(which wget)" ]]; then
  curl -s $C66_CLIENT_URL | tar xz --strip=1
  else
  wget -qO- $C66_CLIENT_URL | tar xz --strip=1
  fi

SCRIPT

  # remind the user to add to $PATH
  if [[ ":$PATH:" != *":/usr/local/cloud66:"* ]]; then
  echo "Add the Cloud 66 CLI to your PATH using:"
  echo "$ echo 'PATH=\"/usr/local/cloud66:\$PATH\"' >> ~/.profile"
  fi

  echo "Installation complete"
}
