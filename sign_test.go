package xorpay

import "testing"

func TestSignFunctions(t *testing.T) {
	if got, want := Sign("abc"), "900150983cd24fb0d6963f7d28e17f72"; got != want {
		t.Fatalf("Sign mismatch: got=%s want=%s", got, want)
	}

	if got, want := SignPay("内容订阅一年期", "native", "50.00", "demo-3", "http://abc.com/xorpay_notify", "mock_secret"), "097a19104039f6f5f20ad0082fad1c14"; got != want {
		t.Fatalf("SignPay mismatch: got=%s want=%s", got, want)
	}

	if got, want := SignQueryByOrderID("demo-order", "secret"), "fc24a300c1ec85c163246e6e4bf4417c"; got != want {
		t.Fatalf("SignQueryByOrderID mismatch: got=%s want=%s", got, want)
	}

	if got, want := SignRefund("0.01", "secret"), "4c70d5a61c87bd3b9709488f37779f05"; got != want {
		t.Fatalf("SignRefund mismatch: got=%s want=%s", got, want)
	}

	if got, want := SignNotify("aoid1", "oid1", "10.00", "1680000000", "secret"), "620e15cab9fec74817011e3eaa80fcbf"; got != want {
		t.Fatalf("SignNotify mismatch: got=%s want=%s", got, want)
	}
}
