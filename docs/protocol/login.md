# Login

# Negotiation

The server must have a valid (can be self signed) x509 certificate to perform handshake. The certificate *MUST* have the `hytale/1` ALPN set, otherwise Dotnet QUIC [will reject the connection](https://learn.microsoft.com/en-us/dotnet/fundamentals/networking/quic/quic-troubleshooting#client-receives-unexpected-alpn-error) (com.hypixel.hytale.server.core.io.transport.QUICTransport).

# Connect

The server will send a connect packet after QUIC handshake is complete.
Here is it's fixed block structure:
```
Offset Size Type         Name                   Notes
0      1    uint8        nullBits               bitfield indicating optional fields
1      64   ascii string protocol hash          fixed length 64 bytes, no prefix
65     1    uint8        client type            enum value
66     16   uuid         uuid                   full 128-bit uuid
82     4    int32 (LE)   language offset        (-1 if none)
86     4    int32 (LE)   identity offset        (-1 if none)
90     4    int32 (LE)   username offset        always present
94     4    int32 (LE)   referral data offset   (-1 if none)
98     4    int32 (LE)   referral source offset (-1 if none)
```

Here is the values of the `nullBits` field, corresponding to the presence of certain variable fields:
```
Bit  Name            
0x01 language present
0x02 identity present
0x04 referral data present
0x08 referral source present
```

Variable fields, read all from the given offset (if present):
- Username = Varstring (always present, ascii)
- Language = Varstring (ascii)
- Identity token = Varstring (utf-8)
- Referral data = byte array (length-prefixed with Varint)
- Referral source = HostAddress