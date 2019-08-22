# Examples of MongoDB Go driver

Requirements:
 * [mongo-go-driver](https://godoc.org/go.mongodb.org/mongo-driver/mongo) v1.1
 * MongoDB v4.0+ 
 * Go v1.11.5+
 

### Custom Encoding/Decoding of Structs 

[custom_encoder_decoder.go](./custom_encoder_decoder.go) file shows how to: 

    * Register a custom encoder/decoder
    * Write a custom EncodeValue
    * Write a custom DecodeValue

The decoder shows how to set a default value when a NULL value is set in the collection. 


