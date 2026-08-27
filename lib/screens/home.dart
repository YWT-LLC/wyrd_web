/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../utils/export.dart';
import '../widgets/export.dart';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:open_ui/open_ui.dart';

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
    return Consumer<EzCP>(
      builder: (_, EzCP config, __) => WyrdWebScaffold(
        config,
        body: EzScreen(
          config,
          child: EzScrollView(
            config,
            children: <Widget>[
              EzHeader(config),

              // Launch //
              // Smoke Signal
              EzElevatedIconButton(
                config,
                onPressed: () async {
                  await ezCmd(
                    'kubectl config use-context docker-desktop -n default',
                    dir: '/Users/mwaldron/repos/flutter/wyrd_web/servers/smoke_signal',
                    onSuccess: doNothing,
                    onFailure: (_) {
                      return;
                    },
                    readout: readout,
                  ); // Needs fix: don't hard-code dir

                  await ezCmd(
                    'skaffold dev',
                    dir: '/Users/mwaldron/repos/flutter/wyrd_web/servers/smoke_signal',
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
              config.divider,

              // Monitor //

              // CLI
              EzRow(
                config,
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
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
                    icon: Icon(showReadout ? Icons.arrow_drop_up : Icons.arrow_drop_down),
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
              config.separator,
            ],
          ),
        ),
        fabs: <Widget>[
          config.spacer,
          SettingsFAB(config, parentContext: context),
        ],
      ),
    );
  }
}
