/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../utils/export.dart';
import '../widgets/export.dart';

import 'package:flutter/material.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
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
    return WyrdWebScaffold(
      EzScreen(EzScrollView(children: <Widget>[
        EzHeader(),

        // Launch //
        // Smoke Signal
        EzElevatedIconButton(
          onPressed: () async {
            await ezCmd(
              'kubectl config use-context docker-desktop -n default',
              dir:
                  '/Users/mwaldron/repos/flutter/wyrd_web/servers/smoke_signal',
              onSuccess: doNothing,
              onFailure: (_) {
                return;
              },
              readout: readout,
            ); // Needs fix: don't hard-code dir

            await ezCmd(
              'skaffold dev',
              dir:
                  '/Users/mwaldron/repos/flutter/wyrd_web/servers/smoke_signal',
              onSuccess: doNothing,
              onFailure: (_) {
                return;
              },
              readout: readout,
            ); // Ditto
          },
          icon: const Icon(Icons.launch),
          label: 'Smoke Signal',
        ),
        const EzDivider(),

        // Monitor //

        // CLI
        EzRow(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            EzText(
              'Console',
              style: EzConfig.styles.titleLarge,
              textAlign: TextAlign.center,
            ),
            EzMargin(vertical: false),
            EzIconButton(
              onPressed: () => setState(() => showReadout = !showReadout),
              icon: Icon(
                showReadout ? Icons.arrow_drop_up : Icons.arrow_drop_down,
              ),
            ),
          ],
        ),
        EzMargin(),

        // Readout
        Visibility(
          visible: showReadout,
          child: Container(
            constraints: BoxConstraints(
              minWidth: widthOf(context) * 0.667,
              maxWidth: widthOf(context) * 0.667,
              maxHeight: heightOf(context) / 2,
            ),
            padding: EdgeInsets.all(EzConfig.marginVal),
            decoration: BoxDecoration(
              color: EzConfig.colors.surfaceDim,
              borderRadius: ezRoundEdge,
            ),
            child: ValueListenableBuilder<String>(
              valueListenable: readout,
              builder: (_, String value, __) => EzScrollView(
                crossAxisAlignment: CrossAxisAlignment.start,
                child: Text(
                  value,
                  style: EzConfig.styles.bodyLarge,
                  textAlign: TextAlign.start,
                ),
              ),
            ),
          ),
        ),
        const EzSeparator(),
      ])),
      fabs: <Widget>[EzConfig.spacer, SettingsFAB(context)],
    );
  }
}
