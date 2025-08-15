package l1

import "fmt"

type PaymentProcessor interface {
	Pay(amount float64) string
}

type OldPaymentSystem struct{}

func (o *OldPaymentSystem) MakePayment(amount int) string {
	return fmt.Sprintf("Paid %d using old system", amount)
}

type PaymentAdapter struct {
	oldSystem *OldPaymentSystem
}

func (p *PaymentAdapter) Pay(amount float64) string {
	return p.oldSystem.MakePayment(int(amount))
}

func Example_L1_21() {

	oldSystem := &OldPaymentSystem{}
	adapter := &PaymentAdapter{oldSystem: oldSystem}

	fmt.Println(adapter.Pay(100500))
}
