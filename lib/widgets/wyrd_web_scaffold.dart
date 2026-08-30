/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import './export.dart';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:open_ui/open_ui.dart';

class WyrdWebScaffold extends StatelessWidget {
  final EzCP config;
  final Widget body;
  final List<Widget>? fabs;
  final bool isHome;

  const WyrdWebScaffold(
    this.config, {
    super.key,
    required this.body,
    this.fabs,
    this.isHome = false,
  });

  @override
  Widget build(BuildContext context) => EzAdaptiveParent(
        small: Consumer<EzCP>(
          builder: (_, EzCP config, __) => EzScaffold(
            config,
            body: body,
            fabs: <Widget>[updater(config), if (fabs != null) ...fabs!, ...config.backFABs(isHome)],
          ),
        ),
      );
}
