Hytale communicates over Quic (UDP).

Note that the server, although it sends data via streams (not dataframes), it uses its own sort of frame format with optional fields, which it does via a "fixed block" of fixed position fields, then a block of fields who's position we can infer from a field offset table. See the connect packet for an example.

A full packet has the format: length (int32), id (int32), data. Note that length is **not** inclusive of ID, this is the full length of the data.