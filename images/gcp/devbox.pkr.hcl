packer {
  required_plugins {
    googlecompute = {
      version = ">= 1.0.0"
      source = "github.com/hashicorp/googlecompute"
    }
  }
}

source "googlecompute" "dev_box" {
  project_id = "gothic-concept-349709"
  source_image = "ubuntu-minimal-2204-jammy-v20220902"
  ssh_username = "root"
  zone = "us-central1-a"
  communicator = "ssh"
  skip_create_image = false
  machine_type = "n2-standard-4"
  image_name = "devbox"
  image_description = "DevBox image with full set of tools for daily coding."
}

build {
  sources = ["sources.googlecompute.dev_box"]
  provisioner "shell" {
    #inline = ["apt -y update && apt -y install jq"]
    script = "../scripts/ubuntu/bootstrap.sh"
  }
}


