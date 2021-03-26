default:
	go build

config:
	go run . gen-config > config.json

send-tx:
	go run . get-balance -address=0x5205f45c6399c41e11e533926ca69a0aedfdbb8d
	go run . get-balance -address=0x5205f45c6399c41e11e533926ca69a0aedfdbb8d

	go run . send-tx -to=0x5205f45c6399c41e11e533926ca69a0aedfdbb8d -value=123

	go run . get-balance -address=0x5205f45c6399c41e11e533926ca69a0aedfdbb8d
	go run . get-balance -address=0x5205f45c6399c41e11e533926ca69a0aedfdbb8d

send-payouts:
	go run . send-payouts

clean:
	-rm zz_*.png
	-rm -rf ./zz_output_address
	-rm -f HayekTool


