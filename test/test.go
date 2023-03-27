package main

type struc struct {
	str string
}

func main() {
	// struc0 := new(struc)
	// struc0.str = "a"

	// struc1 := new(struc)
	// struc1.str = "b"

	// struc2 := new(struc)
	// struc2.str = "c"

	// sli := []*struc{struc0, struc1, struc2}

	// for i := 0; i < len(sli); i++ {
	// 	fmt.Printf("%d - %s   %v\n", i, sli[i], sli)
	// 	if sli[i].str == "a" {
	// 		sli = append(sli[:i], sli[i+1:]...)
	// 		i--
	// 	}
	// }

	// p1 := &pb.Command{CommandId: "client0-1", Key: "ding", Value: "ding"}
	// p2 := &pb.Command{CommandId: "client0-2", Key: "zhang", Value: "han"}
	// // fmt.Printf("%t\n", reflect.DeepEqual(p1, p2))

	// sli := make([]*pb.Command, 1)
	// sli[0] = p1
	// sli2 := make([]*pb.Command, 2)
	// fmt.Printf("%d\n", len(sli2))
	// sli = append(sli, sli2...)
	// sli[2] = p2
	// fmt.Printf("%v\n", sli)
}
