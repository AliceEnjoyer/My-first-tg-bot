package e

import "fmt"

//Wrap - сворачивать

// просто сделали отдельную функцию в отдельном пакете для правильного и удобного сворачивания
func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err) //возвращаем ошибку так, что бы было понятно откуда она произошла
	// + Errorf обарачивает ошибки правильным образом (загугли про errors.Is() и errors.As() - современные способы работы с ошибками в голанг)
}

func WrapIfErr(msg string, err error) error {
	if err != nil {
		return Wrap(msg, err)
	}
	return nil
}
