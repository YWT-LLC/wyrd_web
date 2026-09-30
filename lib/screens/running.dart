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
  const RunningScreen({super.key});

  @override
  State<RunningScreen> createState() => _RunningScreenState();
}

class _RunningScreenState extends State<RunningScreen> {
  // Define the build data //

  ValueNotifier<String> readout = ValueNotifier<String>('');
  bool showReadout = true;

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
          alignment: Alignment.center,
          child:
              EzScrollView(config, mainAxisAlignment: MainAxisAlignment.center, children: <Widget>[
            EzRow(
              config,
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                // Toggle
                EzText(
                  config,
                  text: 'Console',
                  style: config.titleStyle,
                  textAlign: TextAlign.center,
                ),
                config.rowMargin,
                EzIconButton(
                  config,
                  onPressed: () => setState(() => showReadout = !showReadout),
                  tooltip: 'Toggle readout',
                  icon: Icon(showReadout ? Icons.keyboard_arrow_down : Icons.keyboard_arrow_up),
                ),
              ],
            ),
            config.margin,

            // Readout
            Visibility(
              visible: showReadout,
              child: Container(
                constraints: BoxConstraints(
                  minWidth: widthOf(context) * 0.667,
                  maxWidth: widthOf(context) * 0.667,
                  maxHeight: heightOf(context) / 2,
                ),
                padding: EdgeInsets.all(config.marginVal),
                decoration: BoxDecoration(
                  color: config.colors.surfaceDim,
                  borderRadius: config.textRadius,
                ),
                child: ValueListenableBuilder<String>(
                  valueListenable: readout,
                  builder: (_, String value, __) => EzScrollView(
                    config,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    child: Text(value, style: config.bodyStyle, textAlign: TextAlign.start),
                  ),
                ),
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
