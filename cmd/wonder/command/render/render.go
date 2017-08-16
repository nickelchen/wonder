package render

type InfoRender interface {
	Reset(int, int)
	RenderRow(int, string)
	RenderChar(int, int, string, int)

	RenderFlower(int, int)
	RenderTree(int, int)
	RenderGrass(int, int)
	RenderGround(int, int)
	RenderMud(int, int)
}
