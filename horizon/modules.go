package horizon

import (
	"fmt"

	"go.uber.org/fx"
)

var Modules = fx.Module(
	"horizon",
	fx.Provide(
		NewHorizonConfig,
		NewHorizonAuthentication,
		NewHorizonBroadcast,
		NewHorizonCache,
		NewHorizonDatabase,
		NewHorizonLog,
		NewHorizonPrettyJSONEncoder,
		NewHorizonOTP,
		NewHorizonQR,
		NewHorizonRequest,
		NewHorizonSchedule,
		NewHorizonSecurity,
		NewHorizonSMS,
		NewHorizonSMTP,
		NewHorizonStorage,
	),
	fx.Invoke(func(
		request *HorizonRequest,
		log *HorizonLog,
		qr *HorizonQR,
	) {

		encrypted, err := Encrypt("oYrsXzg7eu7Yt5So4e62r7LDVH2hj", "Hello, world!")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Encrypted:", encrypted)

		decrypted, err := Decrypt("oYrsXzg7eu7Yt5So4e62r7LDVH2hj", encrypted)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Decrypted:", decrypted)

		transaction := QRTransaction{
			AccountIDFrom:      "acc1",
			AccountIDTo:        "acc2",
			UserIDFrom:         "user1",
			UserIDTo:           "user2",
			OrganizationIDFrom: "org1",
			OrganizationIDTo:   "org2",
			Amount:             100.5,
		}
		result, err := qr.Encode(transaction)
		fmt.Println(result.QRCode)
		fmt.Println(err)
		fmt.Println("-----")
		decoded, err := qr.Decode(result.QRCode)
		fmt.Println(decoded)
		fmt.Println(err)
		fmt.Println("-----")

		if tx, ok := decoded.(QRTransaction); ok {
			fmt.Println("Amount:", tx.Amount)
			fmt.Println("UserIDFrom:", tx.UserIDFrom)
		} else {
			fmt.Println("Failed to assert type to QRTransaction")
		}
	}),
)
