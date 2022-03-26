## IDATT2104 Onion Router

Navn på løsningen: Tor nexus?

Link til CI/CD:

### Introduksjon

Vi har laget en løsning i GoLang som lar en bruker gjøre en onion routet HTTP-forespørsel. Den blir kryptert tre
ganger før den sendes mellom nodene som dekrypterer ett og ett lag, før den siste noden gjør forespørselen
som brukeren ønsket. Elliptic Curve Diffie-Hellman er brukt for nøkkelutveksling, og Advanced Encryption Standard
med Galois/Counter mode som operasjonsmodus. 

### Implementert funksjonalitet

Når programmet kjøres kan man navigere til localhost:8080 og gjøre onion routede  HTTP-forespørsler. 

### Fremtidig arbeid

Den største mangelen i løsningen vår er at sidene som blir vist ikke er responsive, og å kunne gjøre HTTPS-forespørsler .
### Eksterne avhengigheter

Løsningen er laget utelukkende i native-biblioteket til Go.

### Installasjonsinstrukser

Kun Docker kreves for å kjøre løsningen. https://go.dev/doc/install

### Instruksjoner for å bruke løsningen
1.

### Kjøring av tester

Kjøring av tester kan gjøres ved å navigere til prosjektet og kjøre kommandoen "go test".
