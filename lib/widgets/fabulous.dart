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
  /// [FloatingActionButton] that goes to the [SettingsHomeScreen]
  SettingsFAB(BuildContext context, {super.key})
      : super(
          child: const Icon(Icons.settings),
          onPressed: () => context.goNamed(settingsPath),
          tooltip: l10n.gSettingsHint,
        );
}

const Widget updater = EzUpdaterFAB(
  appVersion: '1.0.0',
  versionSource:
      'https://raw.githubusercontent.com/Empathetech-LLC/wyrd_web/refs/heads/main/APP_VERSION',
  gPlay: 'https://play.google.com/store/apps/details?id=net.empathetech.BLARG',
  appStore: 'https://apps.apple.com/us/app/BLARG/BLARG',
  github: 'https://github.com/Empathetech-LLC/wyrd_web/releases',
);
