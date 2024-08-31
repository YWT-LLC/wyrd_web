import 'lang.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class LangEn extends Lang {
  LangEn([String locale = 'en']) : super(locale);

  @override
  String get gLogoHint =>
      'Empathetic LLC logo: a two dimensional hourglass. Activate to go to the home page';

  @override
  String get gSettingsHint => 'Open the settings page';
}
