#!/usr/bin/env bash

#https://rustup.rs
function install_rust() {
  USERNAME=${1:-"developer"}
  sudo -u $USERNAME -s
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs > ./rustup.sh && chmod +x ./rustup.sh &&  ./rustup.sh -y
}