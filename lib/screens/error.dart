/* smoke_signal
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../widgets/export.dart';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:open_ui/open_ui.dart';

class ErrorScreen extends StatefulWidget {
  const ErrorScreen({super.key});

  @override
  State<ErrorScreen> createState() => _ErrorScreenState();
}

class _ErrorScreenState extends State<ErrorScreen> {
  @override
  Widget build(BuildContext context) => Consumer<EzCP>(
        builder: (_, EzCP config, __) => WyrdWebScaffold(
          config,
          body: EzScreen(
            config,
            alignment: Alignment.center,
            child: EzScrollView(
              config,
              mainAxisAlignment: MainAxisAlignment.center,
              children: <Widget>[
                Text(
                  config.ezL10n.g404Wonder,
                  style: config.headlineStyle,
                  textAlign: TextAlign.center,
                ),
                config.spacer,
                Text(
                  config.ezL10n.g404,
                  style: ezSubTitleStyle(config.styles),
                  textAlign: TextAlign.center,
                ),
                config.separator,
                Text(
                  config.ezL10n.g404Note,
                  style: config.labelStyle,
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
          actions: <HybridAction>[],
        ),
      );
}
