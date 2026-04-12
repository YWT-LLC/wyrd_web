/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import 'package:flutter/material.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

Future<void> install({
  required String dir,
  required void Function() onSuccess,
  required void Function(String) onFailure,
  required ValueNotifier<String> readout,
}) async {
  switch (EzConfig.platform) {
    // Linux //
    case TargetPlatform.linux:
      // Check distro
      await ezCmd(
        'cat /etc/os-release',
        dir: dir,
        onSuccess: doNothing,
        onFailure: onFailure,
        readout: readout,
      );
      final String distro = readout.value.split('\n').last.trim();

      if (distro.contains('Ubuntu') || distro.contains('Debian')) {
        // apt-linux //

        // Install pre-requisites
        await ezCmd(
          'sudo apt-get install gnome-terminal',
          dir: dir,
          onSuccess: doNothing,
          onFailure: onFailure,
          readout: readout,
        );

        // Setup apt
        await ezCmd(
          '''sudo apt-get update -yq && \\
sudo apt-get install -yq ca-certificates curl && \\
sudo install -m 0755 -d /etc/apt/keyrings && \\
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc && \\
sudo chmod a+r /etc/apt/keyrings/docker.asc && \\
echo \\
  "deb [arch=\$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \\
  \$(. /etc/os-release && echo "\${UBUNTU_CODENAME:-\$VERSION_CODENAME}") stable" | \\
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null && \\
sudo apt-get update -yq''',
          dir: dir,
          onSuccess: doNothing,
          onFailure: onFailure,
          readout: readout,
        );

        // Install Docker
        await ezCmd(
          'sudo apt-get install -yq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin',
          dir: dir,
          onSuccess: doNothing,
          onFailure: onFailure,
          readout: readout,
        );

        // Test Docker
        await ezCmd(
          'sudo docker run hello-world',
          dir: dir,
          onSuccess: onSuccess,
          onFailure: onFailure,
          readout: readout,
        );
      } else if (distro.contains('Red Hat')) {
        // RHEL //

        onFailure(
          'RHEL requires a Docker account to run the engine. Please install Docker manually.',
        );
      } else if (distro.contains('Fedora')) {
        // Fedora //

        // Check if the user is running GNOME
        await ezCmd(
          'wmctrl -m',
          dir: dir,
          onSuccess: doNothing,
          onFailure: onFailure,
          readout: readout,
        );

        // Install pre-requisites
        if (readout.value.split('\n').last.trim().contains('GNOME')) {
          await ezCmd(
            '''sudo dnf update -y && \\
sudo dnf install -y gnome-shell-extensions && \\
sudo dnf install -y libappindicator-gtk3 && \\
sudo dnf install -y kstatusnotifieritem-qt && \\
sudo dnf install -y gnome-shell-extension-appindicator && \\
gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com && \\
gnome-extensions enable kstatusnotifieritem-support@gnome-shell-extensions''',
            dir: dir,
            onSuccess: doNothing,
            onFailure: onFailure,
            readout: readout,
          );
        } else {
          await ezCmd(
            'sudo dnf update -y && sudo dnf install gnome-terminal',
            dir: dir,
            onSuccess: doNothing,
            onFailure: onFailure,
            readout: readout,
          );
        }
      } else {
        onFailure('Distro not supported');
      }

      await ezCmd(
        '',
        dir: dir,
        onSuccess: onSuccess,
        onFailure: onFailure,
        readout: readout,
      );
      break;

    // macOS //
    case TargetPlatform.macOS:
      const String version = '184744'; // 4.39.0; 2025-03-05

      // Check for Apple Silicon
      await ezCmd(
        'uname -p',
        dir: dir,
        onSuccess: doNothing,
        onFailure: onFailure,
        readout: readout,
      );

      // Download Docker
      if (readout.value.split('\n').last.trim() == 'arm') {
        await ezCmd(
          'curl https://desktop.docker.com/mac/main/arm64/$version/Docker.dmg',
          dir: dir,
          onSuccess: doNothing,
          onFailure: onFailure,
          readout: readout,
        );
      } else {
        await ezCmd(
          'https://desktop.docker.com/mac/main/amd64/$version/Docker.dmg',
          dir: dir,
          onSuccess: doNothing,
          onFailure: onFailure,
          readout: readout,
        );
      }

      // Install Docker
      await ezCmd(
        '''sudo hdiutil attach Docker.dmg && \\
sudo /Volumes/Docker/Docker.app/Contents/MacOS/install --accept-license && \\
sudo hdiutil detach /Volumes/Docker''',
        dir: dir,
        onSuccess: onSuccess,
        onFailure: onFailure,
        readout: readout,
      );

      break;

    // Windows //
    case TargetPlatform.windows:
      await ezCmd(
        '',
        dir: dir,
        onSuccess: onSuccess,
        onFailure: onFailure,
        readout: readout,
      );
      break;
    default:
      onFailure('Platform not supported');
  }
}
