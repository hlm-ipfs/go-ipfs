package auth

import (
	"net/http"
	"testing"
)

func Test_CreateRetrievalOrder(t *testing.T) {
	header := http.Header{}
	header.Set("Authorization", "Idp eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiMCIsInNjb3BlcyI6WyJyb290Il0sIm1ldGFkYXRhIjp7IjAiOiJhNDQ2NzYzNjg3ZWVkYjQwY2FhMGNiOTdmMWQxZTE2OSIsIjEiOiIwIiwiMTAiOiI0NyIsIjExIjoiYWRtaW4iLCIxMiI6ImFkbWluaXN0cmF0b3IiLCIxMyI6Imh1YW5nZG9uZyIsIjE0IjoiMTIwIiwiMTUiOiIzMTUzNjAwMDAwMDAwMDAwMCIsIjE2IjoiMGZlODI1NGEtMmQ5Yi00NDI5LThmNjAtYWQ1MWE0ZTdkZmU2IiwiMiI6InJvb3QtYXBwIn0sImV4cCI6MTY3MTYyMTgxNCwiaXNzIjoiZ290byIsInN1YiI6IjQ3In0.AajFgvDb5K38s-011reK6uZWtnkjKgBMSwh0Cg1awK9QwlRWUMppXmt_Y90w0nLT1Mj1o1338S4Vn9QoY_DJTleurE6hYSS6S5y4eCORQabAY-0v_FTTUFu-V5lhQjsKhZwsVxU8c4fElFODNcrKDSw6KMJI4JGQFAGy0S5so2o")
	order := ResponseCreateRetrievalOrder{}
	err := CreateRetrievalOrder("QmeKe2J62Xfhcph66HbdvaT1xeCDUEV3wcoy8g3DnDDreW", header, &order)
	if err != nil {
		log.Errorf("respopnse error:%+v", err)
		return
	}
	log.Infof("response:%+v", order)
	VerifyRetrievalToken(order.Token)
}
