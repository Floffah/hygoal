# Handshake

# Negotiation

The server must have a valid (can be self signed) x509 certificate to perform handshake. The certificate *MUST* have the `hytale/1` ALPN set, otherwise Dotnet QUIC [will reject the connection](https://learn.microsoft.com/en-us/dotnet/fundamentals/networking/quic/quic-troubleshooting#client-receives-unexpected-alpn-error) (com.hypixel.hytale.server.core.io.transport.QUICTransport).