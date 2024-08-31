/* wyrd_web
 * Copyright (c) 2024 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../utils/export.dart';
import '../widgets/export.dart';

import 'package:flutter/material.dart';
import 'package:efui_bios/efui_bios.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  // Set the page title //

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    setPageTitle(appTitle);
  }

  // Return the build //

  @override
  Widget build(BuildContext context) {
    return WyrdWebScaffold(
      body: EzScreen(
        margin: EdgeInsets.zero,
        child: Center(
          child: WebOfWyrd(
            size: Size(
              heightOf(context) * (1 / 3),
              heightOf(context) * (2 / 3),
            ),
            margin: 0,
          ),
        ),
      ),
      fab: SettingsFAB(context: context),
    );
  }
}
