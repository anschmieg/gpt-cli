## Projektbeschreibung

Dieses Projekt ist ein portabler **CLI-Wrapper für OpenAI-kompatible LLM APIs**, entwickelt mit **Deno** für eine schlanke, sichere und plattformübergreifende Ausführung.  
Das Tool ermöglicht es, KI-Modelle (z. B. OpenAI, Gemini openai-compatible, Copilot) direkt aus der Kommandozeile anzusprechen – ohne komplexe Installation – und unterstützt konfigurierbare Parameter wie Modellwahl, Temperatur, System-Prompt und Dateiuploads.  

Deno wird aufgrund seiner integrierten Features wie **Permission-Management**, **Bundling zu einer einzelnen Binary** und **Standardbibliotheken** gewählt, um maximale Portabilität und minimale Abhängigkeiten zu gewährleisten.

## Ziele

- **Einfache Nutzung**: Intuitive CLI mit klaren Optionen, ähnlich gängigen Unix-Tools.
- **Portabilität**: Läuft als einzelnes Binary ohne externe Dependencies.
- **Modularität**: Klare Trennung zwischen CLI, Core-Logik und Providern, um spätere Erweiterungen (Streaming, Chat-History, Shell-Suggestions) einfach einzubauen.
- **Multi-Provider-Support**: Einfache Integration verschiedener GPT-kompatibler APIs.
- **Sicherheit**: Striktes Berechtigungsmanagement dank Deno-Permissions.
- **Erweiterbarkeit**: Architektur erlaubt, neue Features schrittweise einzufügen, ohne bestehende Teile stark anzupassen.

Okay – hier ist die Funktionsübersicht mit der dahinterliegenden Logik für die **erste synchrone Version** unseres Deno-CLI-Tools:  

### 1. **CLI‐Parsing (`cli.ts`)**
- **Funktion**: Liest alle Optionen und Argumente (z. B. `--provider`, `--model`, Prompt) ein.  
- **Logik**: Verwendet `std/flags` von Deno, speichert die Werte in einer strukturierten Config. Übergibt Config + Prompt an `core`.

### 2. **Core‐Steuerung (`core.ts`)**
- **Funktion**: Kümmert sich um den Ablauf zwischen CLI und Provider.  
- **Logik**:
  - Nimmt die Config entgegen  
  - Baut daraus den Request (inkl. System-Prompt, Temperatur, max Tokens)  
  - Wählt das passende Provider-Modul  
  - Ruft den synchronen API-Call auf  
  - Gibt das Resultat formatiert zurück (z. B. Plaintext oder Markdown)

### 3. **Provider‐Module (`providers/*.ts`)**
- **Funktion**: Kapseln die API-spezifische Request-Logik (z. B. OpenAI, Gemini).  
- **Logik**:
  - Erstellen den richtigen HTTP-Request (URL, Headers, Body)  
  - Nutzen `fetch` für den API-Call  
  - Extrahieren den relevanten Text aus der JSON-Response  
  - Bei Fehlern → strukturierte Fehlermeldung zurückgeben  

### 4. **Utils (`utils/*.ts`)**
- **Funktion**: Gemeinsame Hilfsfunktionen — z. B. Logging, Clipboard-Integration, Markdown-Rendering.  
- **Logik**:
  - Trennen Präsentationslogik (z. B. Markdown) von API-Logik  
  - Optionales farbiges Debug-Logging, wenn `--verbose`  

### Ablauf gesamt  
`CLI → Config → Core → Provider → fetch → Response → Ausgabe`
