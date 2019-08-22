package main

import (
	"context"
	"reflect"
	"log"
	"time"
	"strings"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MyStruct struct {
	Word    string 
	Number  int64
	Flag    bool
}

func (ms *MyStruct) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	dw, err := vw.WriteDocument()
	if err != nil {
		return err
	}
	for i := 0; i < val.NumField(); i++ {
		fieldKey := val.Type().Field(i).Name
		fieldValue := val.Field(i).Interface()
		//fieldType := val.Field(i).Type().Name()

		vw2, err := dw.WriteDocumentElement(strings.ToLower(fieldKey))
		if err != nil { return err }

		ectx := bsoncodec.EncodeContext{Registry: r.Registry}
		encoder, err := r.LookupEncoder(reflect.TypeOf(fieldValue))
		err = encoder.EncodeValue(ectx, vw2, reflect.ValueOf(fieldValue))
		if err != nil { return err }
	}
	return dw.WriteDocumentEnd()
}
 
func (ms *MyStruct) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	decoderMap := make(map[string]bsoncodec.ValueDecoder, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		fieldKey := val.Type().Field(i).Name
		fieldValue := val.Field(i).Interface() 
		decoder, err := r.LookupDecoder(reflect.TypeOf(fieldValue))
		if err != nil { return err }
		decoderMap[strings.ToLower(fieldKey)] = decoder
	}
	dr, err := vr.ReadDocument()
	if err != nil { return err }
	for {
		name, vr, err := dr.ReadElement()
		if err == bsonrw.ErrEOD { break }
		if err != nil { return err }
		/* Check whether there is a corresponding struct member for the BSON field. If none, skip and don't decode */
		if decoder, ok:= decoderMap[name]; ok {
			dctx := bsoncodec.DecodeContext{Registry: r.Registry}
			name = strings.Title(name)
			/* If the value is NOT NULL */
			if (vr.ReadNull()!=nil) {

				err = decoder.DecodeValue(dctx, vr, val.FieldByName(name))
				if err != nil { 
					log.Println(err) 
					return err
				}
			} else {
				/* Sets default value per type */
				switch val.FieldByName(name).Kind() {
					case reflect.Bool:
						val.FieldByName(name).SetBool(false)
					case reflect.String:
						val.FieldByName(name).SetString("")
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						val.FieldByName(name).SetInt(0)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
						val.FieldByName(name).SetUint(0)
					case reflect.Float32, reflect.Float64:
						val.FieldByName(name).SetFloat(0)
				}
			}
		} else {
			err = vr.Skip()
			if err != nil { return err }
			continue
		}
	}
	return nil
}

func main() {
	mongoURI := "mongodb://localhost:27017/"

	/* Register a custom encoder and decoder for MyStruct */
	reg := bson.NewRegistryBuilder().
			RegisterDefaultEncoder(reflect.Struct, &MyStruct{}).
			RegisterDefaultDecoder(reflect.Struct, &MyStruct{}).
			Build()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI).SetRegistry(reg))
	if err != nil {log.Fatal(err)}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil { log.Fatal(err)}
	collection := client.Database("test").Collection("golang")

	/* Test encode normal values */
	result := MyStruct{}
	doc := MyStruct{Word: "foo", Number: int64(42), Flag: true} 
	responseOne, err := collection.InsertOne(ctx, doc)
	if err != nil { log.Fatal(err) }

	/* Test decode normal values */
    err = collection.FindOne(ctx, bson.D{{"_id", responseOne.InsertedID}}).Decode(&result)
	if err != nil { log.Fatal(err) }
	log.Println(result)

	/* Insert a NULL value string */
	result = MyStruct{}
	responseTwo, err := collection.InsertOne(ctx, bson.M{"word":nil, "number": int64(42)})
	if err != nil { log.Fatal(err) }

	/* Test decode a NULL string value */
	err = collection.FindOne(ctx, bson.D{{"_id", responseTwo.InsertedID}}).Decode(&result)
	if err != nil { log.Fatal(err) }
	log.Println(result)
	
	client.Disconnect(ctx)
}