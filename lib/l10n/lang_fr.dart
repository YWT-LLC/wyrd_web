import 'lang.dart';

// ignore_for_file: type=lint

/// The translations for French (`fr`).
class LangFr extends Lang {
  LangFr([String locale = 'fr']) : super(locale);

  @override
  String get gLogoHint =>
      'Logo de Empathetic LLC : un sablier en deux dimensions. Activer pour accéder à la page d\'accueil';

  @override
  String get gSettingsHint => 'Ouvre la page des paramètres';
}
