/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../screens/export.dart';
import '../utils/export.dart';
import '../widgets/export.dart';
import 'package:ywt_private/ywt_private.dart' as ywt;

import 'package:open_ui/open_ui.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';

class RunningScreen extends StatefulWidget {
  final Services toRun;

  const RunningScreen(this.toRun, {super.key});

  @override
  State<RunningScreen> createState() => _RunningScreenState();
}

class _RunningScreenState extends State<RunningScreen> {
  // Define the build data //

  ValueNotifier<String> readout = ValueNotifier<String>('');

  // Set the page title //

  @override
  void initState() {
    super.initState();
    ezWindowNamer(appName);
  }

  // Return the build //

  @override
  Widget build(BuildContext context) {
    return Consumer<EzCP>(
      builder: (_, EzCP config, __) => WyrdWebScaffold(
        config,
        body: EzScreen(
          config,
          alignment: Alignment.topCenter,
          child: EzScrollView(config, children: <Widget>[
            Container(
              alignment: config.isLTR ? Alignment.topLeft : Alignment.topRight,
              constraints: BoxConstraints(
                minWidth: widthOf(context) * 0.5,
                maxWidth: widthOf(context) * 0.75,
                maxHeight: heightOf(context) * 0.75,
              ),
              child: ExpansionTile(
                onExpansionChanged: (_) => setState(() {}),
                title: EzText(
                  config,
                  text: 'Console',
                  style: config.titleStyle,
                  textAlign: TextAlign.center,
                ),
                children: <Widget>[
                  EzTextBackground(
                    config,
                    text: ValueListenableBuilder<String>(
                      valueListenable: readout,
                      builder: (_, String value, __) => Text(
                        value,
                        style: config.bodyStyle,
                        textAlign: TextAlign.start,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Footer
            EzFooter(config, a11howPath: ywt.wyrdWebContributeA11),
          ]),
        ),
        isHome: true,
        actions: <HybridAction>[
          HybridAction(
            label: config.ezL10n.gSettings,
            icon: Icons.settings,
            onPressed: () => context.goNamed(settingsHubPath),
          ),
        ],
      ),
    );
  }
}
