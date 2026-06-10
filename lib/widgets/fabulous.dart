/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../screens/export.dart';
import '../utils/export.dart';

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class SettingsFAB extends FloatingActionButton {
  final EzCP config;

  SettingsFAB(this.config, {required BuildContext parentContext, super.key})
      : super(
          child: EzIcon(config, Icons.settings),
          onPressed: () => parentContext.goNamed(settingsHubPath),
          tooltip: l10n(config).gSettingsHint,
        );
}

EzUpdaterFAB updater(EzCP config) => EzUpdaterFAB(
      config,
      appVersion: '1.0.0',
      versionSource:
          'https://raw.githubusercontent.com/Empathetech-LLC/wyrd_web/refs/heads/main/APP_VERSION',
      gPlay: 'https://play.google.com/store/apps/details?id=net.empathetech.BLARG',
      appStore: 'https://apps.apple.com/us/app/BLARG/BLARG',
      github: 'https://github.com/Empathetech-LLC/wyrd_web/releases',
    );
