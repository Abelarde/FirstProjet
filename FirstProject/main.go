package main

import (
	"fmt"
	archivos "github.com/Abelarde/FirstProject/lib"
	"strings"
)

//TODO: RECORDAR DE NO ESPACIOS EN \*
//TODO: QUIZAS QUE EN CONSOLA ACEPTE MAS DE UN \*
//TODO: OJO EL FIT DE LAS EXTENDIDAS SIEMPRE TIENE QUE SER PRIMER AJUSTE PORQUE SUS LOGICAS SE CREAN CON EL PRIMER AJUSTE

//TODO: VER QUE VALOR LE ESTOY PASANDO EN EL FIT
//TODO: REVISAR LA LOGICA UNIDA

//TODO: TERMINAR DE IMPLEMENTAR EL DELETE DE FDISK CON LOS METODS QUE YA TENGO (HICE EN MKFS)

func main() {
	fmt.Print("================================================================\n")
	fmt.Print("===========		JOSUE EDUARDO ABELARDE PEREZ  ==========\n")
	fmt.Print("===========	             SISTEMA DE ARCHIVOS      ==========\n")
	fmt.Print("===========		           USAC		      ==========\n")
	fmt.Print("================================================================\n")

	for {
		fmt.Println("======================================")
		fmt.Println("Please enter a command:")
		entrada := archivos.CadenaConsola()
		if strings.Contains(entrada, "\\*") {
			entrada = entrada + archivos.CadenaConsola()
		}
		fmt.Println(entrada)
		fmt.Println("======================================")
		archivos.Separar(entrada)
	}

	//archivos.CrearArchivo()
	//archivos.Prue()

}

//func readFile() {
//
//	file, err := os.Open("test.dsk")
//	defer file.Close()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	m := payload{}
//	for i := 0; i < 10; i++ {
//		data := readBytes(file, 16)
//		buffer := bytes.NewBuffer(data)
//		err = binary.Read(buffer, binary.BigEndian, &m)
//		if err != nil {
//			log.Fatal("binary.Read failed", err)
//		}
//
//		fmt.Println(m)
//	}
//
//}
//
//func readBytes(file *os.File, number int) []byte {
//	bytes := make([]byte, number)
//
//	_, err := file.Read(bytes)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return bytes
//}

//HasSuffix

//ComandoList := archivos.ComandoExec()
//archivos.PrintSlice("ComandoList", ComandoList)
/*
mkdisk -SiZe->8 -pAth->”/home/mis discos/DISCO Prueba/” \*
-namE->Disco_3.dsk

Mkdisk -size->16 -path->”/home/mis discos/” -NaMe->Disco4.dsk

rmDisk -path->"/home/mis discos/Disco_4.dsk"

Fdisk -sizE->72 -path->/home/Disco1.dsk -name->Particion1
Mkdisk -size->16 -path->/home/misdiscos/ -NaMe->Disco4.dsk
Mkdisk -size->5 -path->/home/user/ -name->Hoja1_201602890.dsk -uniT->M
rmDisk -path->/home/misdiscos/Disco_4.dsk

unmount -id1->vda1 -id2->vdb2 -id3->vdc1

comando3\*comando4

ESTOS ME DAN ERROR:
fdisk -sizE->72 -path->/home/Disco1.dsk -name->Particion1 -name->Particion2
Mkdisk -size->16 -path->”/home/mis discos/” -NaMe->Disco4.dsk
Mkdisk -size->16 -path->/home/misdiscos/ -NaMe->Disco4.dsk
Mkdisk -size->16 -path->/home/misdiscos/misdiscos2/misdiscos3/misdiscos4/ -NaMe->Disco4.dsk


*/

/*
" " \* _  combinados normal

//cadena directa [2 ejemplos]
Mkdisk -size->16 -path->/home/misdiscos/ -NaMe->Disco4.dsk
rmDisk -path->/home/misdiscos/Disco_4.dsk

//combinado \* y " "
mkdisk -SiZe->8 -pAth->"/home/mis discos/DISCO Prueba/"\*
-namE->Disco_3.dsk
mkdisk -SiZe->8 -pAth->"/home/mis discos/Prueba/" \*
-namE->Disco_3.dsk
mkdisk -SiZe->8 -pAth->"/home/mis discos/" \*
-namE->Disco_3.dsk

//cadena con \* [2 ejemplos]
Mkdisk -size->32 -path->/home/user/\*
-name->Disco1.dsk -uniT->k
Mkdisk -size->32 -path->/home/user/ \*
-name->Disco_1.dsk -uniT->k

//con " "
Mkdisk -size->16 -path->/home/misdiscos/ -NaMe->Disco4.dsk
Mkdisk -size->16 -path->"/home/mi espacio/" -NaMe->Disco_4.dsk
Mkdisk -size->16 -path->/home/misdiscos/ -NaMe->Disco4.dsk
Mkdisk -size->16 -path->"/home/mi espacio/" -NaMe->Disco_4.dsk

Mkdisk -size->16 -path->/home/misdiscos/ -NaMe->Disco5.dsk
Mkdisk -size->16 -path->"/home/mi espacio/" -NaMe->Disco_5.dsk

Mkdisk -size->16 -path->/home/user/ -NaMe->Disco5.dsk
Fdisk -sizE->7 -path->/home/user/Disco5.dsk -name->Particion1
Fdisk -sizE->3 -path->/home/user/Disco5.dsk -type->L -name->Particion3
Fdisk -sizE->1 -path->/home/user/Disco_6.dsk -type->p -name->Particion3

mount -path->/home/user/Disco_6.dsk -name->Particion1
mount -path->/home/user/Disco5.dsk -name->Particion3
mount -path->/home/user/Disco_6.dsk -name->Particion3

mount -path->/home/user/Disco5.dsk -name->Particion3
rep -id->vda1 -Path->/home/user/reports/reporte1.jpg -name->disk

TODO: VER SI LAS LETRAS NO ME DA PROBLEMAS TAMBIEN


unmount -id1->vda1

exec -path->/home/eduardo/go/src/github.com/Abelarde/Proyecto1/Entrada.mia
exec -path->/home/eduardo/go/src/github.com/Abelarde/FirstProject/Entrada.mia

exec -path->/home/eduardo/go/src/github.com/Abelarde/FirstProject/auxiliar.mia

exec -path->/home/eduardo/go/src/github.com/Abelarde/FirstProject/auxiliarDos.mia

Fdisk -sizE->72 -path->/home/user/Disco5.dsk -name->Particion1
Fdisk -sizE->72 -path->/home/user/Disco5.dsk -name->Particion1

*/
