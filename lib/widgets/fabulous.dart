/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../screens/export.dart';
import '../utils/export.dart';

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:open_ui/open_ui.dart';

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
      appVersion: '1.0.1',
      versionSource:
          'https://raw.githubusercontent.com/YWT-LLC/wyrd_web/refs/heads/main/APP_VERSION',
      gPlay: 'https://play.google.com/store/apps/details?id=llc.ywt.wyrd_web',
      appStore: 'https://apps.apple.com/us/app/BLARG/BLARG',
      github: 'https://github.com/YWT-LLC/wyrd_web/releases',
    );
